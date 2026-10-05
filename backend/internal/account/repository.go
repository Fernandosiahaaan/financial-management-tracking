package account

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("an account with this name already exists")
)

// Repository defines data operations for accounts.
type Repository interface {
	Create(ctx context.Context, acc Account) (*Account, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Account, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Account, error)
	Update(ctx context.Context, acc Account) (*Account, error)
	SoftDelete(ctx context.Context, userID, id uuid.UUID) error
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create inserts a new account scoped to the user.
func (r *PostgresRepository) Create(ctx context.Context, acc Account) (*Account, error) {
	query := `
		INSERT INTO accounts (user_id, name, type, opening_balance, current_balance, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, name, type, opening_balance, current_balance, status, deleted_at, created_at, updated_at;
	`

	cleanName := strings.TrimSpace(acc.Name)
	if acc.Status == "" {
		acc.Status = StatusActive
	}
	// Initial current_balance starts at opening_balance when created
	if acc.CurrentBalance == 0 && acc.OpeningBalance != 0 {
		acc.CurrentBalance = acc.OpeningBalance
	}

	var created Account
	err := r.pool.QueryRow(ctx, query,
		acc.UserID,
		cleanName,
		acc.Type,
		acc.OpeningBalance,
		acc.CurrentBalance,
		acc.Status,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Name,
		&created.Type,
		&created.OpeningBalance,
		&created.CurrentBalance,
		&created.Status,
		&created.DeletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAccountAlreadyExists
		}
		return nil, fmt.Errorf("create account: %w", err)
	}

	created.FormattedBalance = money.FormatIDR(created.CurrentBalance)
	return &created, nil
}

// ListByUser returns all active accounts belonging to the user.
func (r *PostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	query := `
		SELECT id, user_id, name, type, opening_balance, current_balance, status, deleted_at, created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(
			&acc.ID,
			&acc.UserID,
			&acc.Name,
			&acc.Type,
			&acc.OpeningBalance,
			&acc.CurrentBalance,
			&acc.Status,
			&acc.DeletedAt,
			&acc.CreatedAt,
			&acc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		acc.FormattedBalance = money.FormatIDR(acc.CurrentBalance)
		accounts = append(accounts, acc)
	}

	if accounts == nil {
		accounts = []Account{}
	}

	return accounts, nil
}

// GetByID returns an active account by its ID and UserID.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Account, error) {
	query := `
		SELECT id, user_id, name, type, opening_balance, current_balance, status, deleted_at, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
	`

	var acc Account
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Name,
		&acc.Type,
		&acc.OpeningBalance,
		&acc.CurrentBalance,
		&acc.Status,
		&acc.DeletedAt,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("get account by id: %w", err)
	}

	acc.FormattedBalance = money.FormatIDR(acc.CurrentBalance)
	return &acc, nil
}

// Update modifies an account's name, type, or status.
func (r *PostgresRepository) Update(ctx context.Context, acc Account) (*Account, error) {
	query := `
		UPDATE accounts
		SET name = $3, type = $4, status = $5, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, name, type, opening_balance, current_balance, status, deleted_at, created_at, updated_at;
	`

	cleanName := strings.TrimSpace(acc.Name)
	var updated Account
	err := r.pool.QueryRow(ctx, query,
		acc.ID,
		acc.UserID,
		cleanName,
		acc.Type,
		acc.Status,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Name,
		&updated.Type,
		&updated.OpeningBalance,
		&updated.CurrentBalance,
		&updated.Status,
		&updated.DeletedAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAccountAlreadyExists
		}
		return nil, fmt.Errorf("update account: %w", err)
	}

	updated.FormattedBalance = money.FormatIDR(updated.CurrentBalance)
	return &updated, nil
}

// SoftDelete sets deleted_at to the current timestamp.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	query := `
		UPDATE accounts
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
	`

	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	return nil
}
