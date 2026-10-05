package budget

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access methods for budgets.
type Repository interface {
	Create(ctx context.Context, b Budget) (*Budget, error)
	ListByUserAndCycle(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]Budget, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Budget, error)
	Update(ctx context.Context, userID, id uuid.UUID, plannedAmount int64) (*Budget, error)
	SoftDelete(ctx context.Context, userID, id uuid.UUID) error
	GetActualSpending(ctx context.Context, userID, categoryID uuid.UUID, cycleStart, cycleEnd string) (int64, error)
	CategoryBelongsToUser(ctx context.Context, userID, categoryID uuid.UUID) (bool, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository for budgets.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create inserts a new budget scoped to the user.
func (r *PostgresRepository) Create(ctx context.Context, b Budget) (*Budget, error) {
	query := `
		INSERT INTO budgets (user_id, category_id, cycle_start, cycle_end, planned_amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, category_id, cycle_start::text, cycle_end::text, planned_amount, created_at, updated_at, deleted_at;
	`

	var created Budget
	err := r.pool.QueryRow(ctx, query,
		b.UserID,
		b.CategoryID,
		b.CycleStart,
		b.CycleEnd,
		b.PlannedAmount,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.CategoryID,
		&created.CycleStart,
		&created.CycleEnd,
		&created.PlannedAmount,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrBudgetAlreadyExists
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("create budget: %w", err)
	}

	return &created, nil
}

// ListByUserAndCycle returns all active budgets for a given user and cycle, joined with category info.
func (r *PostgresRepository) ListByUserAndCycle(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]Budget, error) {
	query := `
		SELECT b.id, b.user_id, b.category_id, b.cycle_start::text, b.cycle_end::text, b.planned_amount,
		       b.created_at, b.updated_at, b.deleted_at,
		       c.name AS category_name, c.icon AS category_icon, c.color AS category_color
		FROM budgets b
		JOIN categories c ON c.id = b.category_id AND c.deleted_at IS NULL
		WHERE b.user_id = $1
		  AND b.cycle_start = $2
		  AND b.cycle_end = $3
		  AND b.deleted_at IS NULL
		ORDER BY c.name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID, cycleStart, cycleEnd)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var b Budget
		err := rows.Scan(
			&b.ID, &b.UserID, &b.CategoryID,
			&b.CycleStart, &b.CycleEnd, &b.PlannedAmount,
			&b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
			&b.CategoryName, &b.CategoryIcon, &b.CategoryColor,
		)
		if err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		budgets = append(budgets, b)
	}

	if budgets == nil {
		budgets = []Budget{}
	}

	return budgets, nil
}

// GetByID finds an active budget by its ID and UserID.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Budget, error) {
	query := `
		SELECT b.id, b.user_id, b.category_id, b.cycle_start::text, b.cycle_end::text, b.planned_amount,
		       b.created_at, b.updated_at, b.deleted_at,
		       c.name AS category_name, c.icon AS category_icon, c.color AS category_color
		FROM budgets b
		JOIN categories c ON c.id = b.category_id AND c.deleted_at IS NULL
		WHERE b.id = $1 AND b.user_id = $2 AND b.deleted_at IS NULL;
	`

	var b Budget
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&b.ID, &b.UserID, &b.CategoryID,
		&b.CycleStart, &b.CycleEnd, &b.PlannedAmount,
		&b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		&b.CategoryName, &b.CategoryIcon, &b.CategoryColor,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBudgetNotFound
		}
		return nil, fmt.Errorf("get budget by id: %w", err)
	}

	return &b, nil
}

// Update modifies the planned amount of an existing budget.
func (r *PostgresRepository) Update(ctx context.Context, userID, id uuid.UUID, plannedAmount int64) (*Budget, error) {
	query := `
		UPDATE budgets
		SET planned_amount = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, category_id, cycle_start::text, cycle_end::text, planned_amount, created_at, updated_at, deleted_at;
	`

	var b Budget
	err := r.pool.QueryRow(ctx, query, id, userID, plannedAmount).Scan(
		&b.ID, &b.UserID, &b.CategoryID,
		&b.CycleStart, &b.CycleEnd, &b.PlannedAmount,
		&b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBudgetNotFound
		}
		return nil, fmt.Errorf("update budget: %w", err)
	}

	return &b, nil
}

// SoftDelete sets deleted_at on a budget.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	query := `
		UPDATE budgets
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
	`

	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete budget: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBudgetNotFound
	}

	return nil
}

// GetActualSpending calculates the total EXPENSE transactions for a specific category within a cycle.
func (r *PostgresRepository) GetActualSpending(ctx context.Context, userID, categoryID uuid.UUID, cycleStart, cycleEnd string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND category_id = $2
		  AND type = 'EXPENSE'
		  AND transaction_date >= $3
		  AND transaction_date <= $4
		  AND deleted_at IS NULL;
	`

	var total int64
	err := r.pool.QueryRow(ctx, query, userID, categoryID, cycleStart, cycleEnd).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get actual spending: %w", err)
	}

	return total, nil
}

// CategoryBelongsToUser checks if a category exists and belongs to the given user.
func (r *PostgresRepository) CategoryBelongsToUser(ctx context.Context, userID, categoryID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM categories
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		);
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, categoryID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check category ownership: %w", err)
	}

	return exists, nil
}
