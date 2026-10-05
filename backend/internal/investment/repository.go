package investment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Repository defines persistence operations for investments.
type Repository interface {
	Create(ctx context.Context, inv Investment) (*Investment, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Investment, error)
	List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Investment, error)
	UpdateValuation(ctx context.Context, userID, id uuid.UUID, currentValue int64, notes *string) (*Investment, error)
	SoftDelete(ctx context.Context, userID, id uuid.UUID) error
	GetTotalInvestmentsValue(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

func formatInvestment(inv *Investment) {
	if inv == nil {
		return
	}
	inv.FormattedCapital = money.FormatIDR(inv.Capital)
	inv.FormattedCurrentValue = money.FormatIDR(inv.CurrentValue)
	inv.FormattedUnrealizedGain = money.FormatIDR(inv.UnrealizedGain)
}

// NewRepository creates a new PostgresRepository.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create inserts a new investment holding, optionally deducting capital from a source account.
func (r *PostgresRepository) Create(ctx context.Context, inv Investment) (*Investment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var sourceAccName string
	if inv.SourceAccountID != nil && inv.Capital > 0 {
		var currentBalance int64
		err := tx.QueryRow(ctx, `
			SELECT name, current_balance
			FROM accounts
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
			FOR UPDATE
		`, *inv.SourceAccountID, inv.UserID).Scan(&sourceAccName, &currentBalance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrAccountNotFound
			}
			return nil, fmt.Errorf("lookup source account: %w", err)
		}

		if currentBalance < inv.Capital {
			return nil, ErrInsufficientBalance
		}

		_, err = tx.Exec(ctx, `
			UPDATE accounts
			SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, inv.Capital, *inv.SourceAccountID, inv.UserID)
		if err != nil {
			return nil, fmt.Errorf("deduct investment capital: %w", err)
		}
	}

	query := `
		INSERT INTO investments (
			user_id, source_account_id, type, name, capital, current_value, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id, created_at, updated_at;
	`

	created := inv
	created.SourceAccountName = sourceAccName
	created.UnrealizedGain = inv.CurrentValue - inv.Capital
	if inv.Capital > 0 {
		created.UnrealizedGainPercentage = (float64(created.UnrealizedGain) / float64(inv.Capital)) * 100
	}

	err = tx.QueryRow(ctx, query,
		inv.UserID,
		inv.SourceAccountID,
		inv.Type,
		inv.Name,
		inv.Capital,
		inv.CurrentValue,
		inv.Notes,
	).Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert investment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit investment transaction: %w", err)
	}

	return &created, nil
}

// GetByID returns an investment holding by ID.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Investment, error) {
	query := `
		SELECT
			i.id, i.user_id, i.source_account_id, COALESCE(a.name, ''),
			i.type, i.name, i.capital, i.current_value, i.notes,
			i.created_at, i.updated_at
		FROM investments i
		LEFT JOIN accounts a ON a.id = i.source_account_id
		WHERE i.id = $1 AND i.user_id = $2 AND i.deleted_at IS NULL;
	`

	var inv Investment
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&inv.ID,
		&inv.UserID,
		&inv.SourceAccountID,
		&inv.SourceAccountName,
		&inv.Type,
		&inv.Name,
		&inv.Capital,
		&inv.CurrentValue,
		&inv.Notes,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvestmentNotFound
		}
		return nil, fmt.Errorf("query investment: %w", err)
	}

	inv.UnrealizedGain = inv.CurrentValue - inv.Capital
	if inv.Capital > 0 {
		inv.UnrealizedGainPercentage = (float64(inv.UnrealizedGain) / float64(inv.Capital)) * 100
	}
	formatInvestment(&inv)

	return &inv, nil
}

// List returns user investments with optional type filter.
func (r *PostgresRepository) List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Investment, error) {
	query := `
		SELECT
			i.id, i.user_id, i.source_account_id, COALESCE(a.name, ''),
			i.type, i.name, i.capital, i.current_value, i.notes,
			i.created_at, i.updated_at
		FROM investments i
		LEFT JOIN accounts a ON a.id = i.source_account_id
		WHERE i.user_id = $1 AND i.deleted_at IS NULL
	`
	args := []interface{}{userID}

	if filter.Type != nil {
		args = append(args, *filter.Type)
		query += fmt.Sprintf(" AND i.type = $%d", len(args))
	}

	query += " ORDER BY i.created_at DESC;"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list investments: %w", err)
	}
	defer rows.Close()

	var investments []Investment
	for rows.Next() {
		var inv Investment
		err := rows.Scan(
			&inv.ID,
			&inv.UserID,
			&inv.SourceAccountID,
			&inv.SourceAccountName,
			&inv.Type,
			&inv.Name,
			&inv.Capital,
			&inv.CurrentValue,
			&inv.Notes,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan investment: %w", err)
		}
		inv.UnrealizedGain = inv.CurrentValue - inv.Capital
		if inv.Capital > 0 {
			inv.UnrealizedGainPercentage = (float64(inv.UnrealizedGain) / float64(inv.Capital)) * 100
		}
		investments = append(investments, inv)
	}

	if investments == nil {
		investments = []Investment{}
	}
	for i := range investments {
		formatInvestment(&investments[i])
	}

	return investments, nil
}

// UpdateValuation updates the current value of an investment.
func (r *PostgresRepository) UpdateValuation(ctx context.Context, userID, id uuid.UUID, currentValue int64, notes *string) (*Investment, error) {
	query := `
		UPDATE investments
		SET current_value = $1,
		    notes = CASE WHEN $2::text IS NOT NULL THEN $2 ELSE notes END,
		    updated_at = NOW()
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
		RETURNING id;
	`
	var returnedID uuid.UUID
	err := r.pool.QueryRow(ctx, query, currentValue, notes, id, userID).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvestmentNotFound
		}
		return nil, fmt.Errorf("update investment valuation: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

// SoftDelete marks an investment as deleted.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE investments
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete investment: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrInvestmentNotFound
	}
	return nil
}

// GetTotalInvestmentsValue computes total current market valuation across user's investments.
func (r *PostgresRepository) GetTotalInvestmentsValue(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(current_value), 0)
		FROM investments
		WHERE user_id = $1 AND deleted_at IS NULL;
	`
	var total int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("compute total investments value: %w", err)
	}
	return total, nil
}
