package allocation

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access methods for allocations.
type Repository interface {
	Create(ctx context.Context, a Allocation) (*Allocation, error)
	ListByUserAndCycle(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]Allocation, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Allocation, error)
	Update(ctx context.Context, userID, id uuid.UUID, allocatedAmount int64) (*Allocation, error)
	SoftDelete(ctx context.Context, userID, id uuid.UUID) error
	GetTotalAllocated(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string, excludeID *uuid.UUID) (int64, error)
	GetCycleIncome(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) (int64, error)
	CategoryBelongsToUser(ctx context.Context, userID, categoryID uuid.UUID) (bool, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository for allocations.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create inserts a new allocation scoped to the user.
func (r *PostgresRepository) Create(ctx context.Context, a Allocation) (*Allocation, error) {
	query := `
		INSERT INTO allocations (user_id, category_id, cycle_start, cycle_end, allocated_amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, category_id, cycle_start::text, cycle_end::text, allocated_amount, created_at, updated_at, deleted_at;
	`

	var created Allocation
	err := r.pool.QueryRow(ctx, query,
		a.UserID,
		a.CategoryID,
		a.CycleStart,
		a.CycleEnd,
		a.AllocatedAmount,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.CategoryID,
		&created.CycleStart,
		&created.CycleEnd,
		&created.AllocatedAmount,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAllocationAlreadyExists
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("create allocation: %w", err)
	}

	return &created, nil
}

// ListByUserAndCycle returns all active allocations for a given user and cycle, joined with category info.
func (r *PostgresRepository) ListByUserAndCycle(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]Allocation, error) {
	query := `
		SELECT a.id, a.user_id, a.category_id, a.cycle_start::text, a.cycle_end::text, a.allocated_amount,
		       a.created_at, a.updated_at, a.deleted_at,
		       c.name AS category_name, c.icon AS category_icon, c.color AS category_color
		FROM allocations a
		JOIN categories c ON c.id = a.category_id AND c.deleted_at IS NULL
		WHERE a.user_id = $1
		  AND a.cycle_start = $2
		  AND a.cycle_end = $3
		  AND a.deleted_at IS NULL
		ORDER BY c.name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID, cycleStart, cycleEnd)
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	defer rows.Close()

	var allocations []Allocation
	for rows.Next() {
		var a Allocation
		err := rows.Scan(
			&a.ID, &a.UserID, &a.CategoryID,
			&a.CycleStart, &a.CycleEnd, &a.AllocatedAmount,
			&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
			&a.CategoryName, &a.CategoryIcon, &a.CategoryColor,
		)
		if err != nil {
			return nil, fmt.Errorf("scan allocation: %w", err)
		}
		allocations = append(allocations, a)
	}

	if allocations == nil {
		allocations = []Allocation{}
	}

	return allocations, nil
}

// GetByID finds an active allocation by its ID and UserID.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Allocation, error) {
	query := `
		SELECT a.id, a.user_id, a.category_id, a.cycle_start::text, a.cycle_end::text, a.allocated_amount,
		       a.created_at, a.updated_at, a.deleted_at,
		       c.name AS category_name, c.icon AS category_icon, c.color AS category_color
		FROM allocations a
		JOIN categories c ON c.id = a.category_id AND c.deleted_at IS NULL
		WHERE a.id = $1 AND a.user_id = $2 AND a.deleted_at IS NULL;
	`

	var a Allocation
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&a.ID, &a.UserID, &a.CategoryID,
		&a.CycleStart, &a.CycleEnd, &a.AllocatedAmount,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
		&a.CategoryName, &a.CategoryIcon, &a.CategoryColor,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAllocationNotFound
		}
		return nil, fmt.Errorf("get allocation by id: %w", err)
	}

	return &a, nil
}

// Update modifies the allocated amount of an existing allocation.
func (r *PostgresRepository) Update(ctx context.Context, userID, id uuid.UUID, allocatedAmount int64) (*Allocation, error) {
	query := `
		UPDATE allocations
		SET allocated_amount = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, category_id, cycle_start::text, cycle_end::text, allocated_amount, created_at, updated_at, deleted_at;
	`

	var a Allocation
	err := r.pool.QueryRow(ctx, query, id, userID, allocatedAmount).Scan(
		&a.ID, &a.UserID, &a.CategoryID,
		&a.CycleStart, &a.CycleEnd, &a.AllocatedAmount,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAllocationNotFound
		}
		return nil, fmt.Errorf("update allocation: %w", err)
	}

	return &a, nil
}

// SoftDelete sets deleted_at on an allocation.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	query := `
		UPDATE allocations
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
	`

	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete allocation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAllocationNotFound
	}

	return nil
}

// GetTotalAllocated calculates the sum of all active allocations for a cycle, optionally excluding one allocation.
func (r *PostgresRepository) GetTotalAllocated(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string, excludeID *uuid.UUID) (int64, error) {
	var query string
	var args []interface{}

	if excludeID != nil {
		query = `
			SELECT COALESCE(SUM(allocated_amount), 0)
			FROM allocations
			WHERE user_id = $1
			  AND cycle_start = $2
			  AND cycle_end = $3
			  AND id != $4
			  AND deleted_at IS NULL;
		`
		args = []interface{}{userID, cycleStart, cycleEnd, *excludeID}
	} else {
		query = `
			SELECT COALESCE(SUM(allocated_amount), 0)
			FROM allocations
			WHERE user_id = $1
			  AND cycle_start = $2
			  AND cycle_end = $3
			  AND deleted_at IS NULL;
		`
		args = []interface{}{userID, cycleStart, cycleEnd}
	}

	var total int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get total allocated: %w", err)
	}

	return total, nil
}

// GetCycleIncome calculates total INCOME transactions within the cycle.
func (r *PostgresRepository) GetCycleIncome(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		  AND type = 'INCOME'
		  AND transaction_date >= $2
		  AND transaction_date <= $3
		  AND deleted_at IS NULL;
	`

	var total int64
	err := r.pool.QueryRow(ctx, query, userID, cycleStart, cycleEnd).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get cycle income: %w", err)
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
