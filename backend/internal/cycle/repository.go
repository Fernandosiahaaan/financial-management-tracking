package cycle

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access methods for financial cycle calculations and aggregations.
type Repository interface {
	GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error)
	GetCycleAggregates(ctx context.Context, userID uuid.UUID, startDate, endDate string) (int64, int64, int, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository for financial cycles.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetUserCycleStartDay retrieves the user's preferred cycle start day (1-31) from user_settings.
func (r *PostgresRepository) GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT cycle_start_day
		FROM user_settings
		WHERE user_id = $1;
	`

	var startDay int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&startDay)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Default to 1 if no settings row found yet
			return 1, nil
		}
		return 1, fmt.Errorf("get user cycle_start_day: %w", err)
	}

	if startDay < 1 || startDay > 31 {
		return 1, nil
	}

	return startDay, nil
}

// GetCycleAggregates queries transactions strictly within [startDate, endDate] for the user,
// calculating total income, total expense, and transaction count.
// Invariant: Transfers are not expenses and do not alter operational income/expense.
func (r *PostgresRepository) GetCycleAggregates(ctx context.Context, userID uuid.UUID, startDate, endDate string) (int64, int64, int, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS total_expense,
			COUNT(*) AS total_count
		FROM transactions
		WHERE user_id = $1 
		  AND transaction_date >= $2 
		  AND transaction_date <= $3 
		  AND deleted_at IS NULL;
	`

	var totalIncome, totalExpense int64
	var count int

	err := r.pool.QueryRow(ctx, query, userID, startDate, endDate).Scan(
		&totalIncome,
		&totalExpense,
		&count,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("get cycle aggregates: %w", err)
	}

	return totalIncome, totalExpense, count, nil
}
