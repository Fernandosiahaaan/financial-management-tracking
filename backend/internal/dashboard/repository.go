package dashboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data aggregation operations for the dashboard.
type Repository interface {
	GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error)
	GetCycleCashFlow(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) (income, expense int64, err error)
	GetBudgetProgress(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]BudgetProgress, error)
	GetAccountSummaries(ctx context.Context, userID uuid.UUID) ([]AccountSummary, int64, error)
	GetReceivablesAndInvestmentsTotals(ctx context.Context, userID uuid.UUID) (totalReceivables, totalInvestments int64, err error)
	GetTotalAssetsAsOf(ctx context.Context, userID uuid.UUID, asOfDate string) (int64, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository for dashboard.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetUserCycleStartDay retrieves the user's configured cycle start day.
func (r *PostgresRepository) GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT cycle_start_day FROM user_settings WHERE user_id = $1;`
	var startDay int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&startDay)
	if err != nil {
		return 1, nil // Default to day 1 if not set
	}
	return startDay, nil
}

// GetCycleCashFlow calculates total income and expense within the specified cycle window.
func (r *PostgresRepository) GetCycleCashFlow(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) (int64, int64, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE user_id = $1
		  AND transaction_date >= $2
		  AND transaction_date <= $3
		  AND deleted_at IS NULL;
	`

	var income, expense int64
	err := r.pool.QueryRow(ctx, query, userID, cycleStart, cycleEnd).Scan(&income, &expense)
	if err != nil {
		return 0, 0, fmt.Errorf("get cycle cash flow: %w", err)
	}

	return income, expense, nil
}

// GetBudgetProgress returns active budgets for the cycle with actual spending and variance computed from transactions.
func (r *PostgresRepository) GetBudgetProgress(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]BudgetProgress, error) {
	query := `
		SELECT
			b.category_id,
			c.name AS category_name,
			c.icon AS category_icon,
			c.color AS category_color,
			b.planned_amount,
			COALESCE(t.actual_amount, 0) AS actual_amount
		FROM budgets b
		JOIN categories c ON c.id = b.category_id AND c.deleted_at IS NULL
		LEFT JOIN (
			SELECT category_id, SUM(amount) AS actual_amount
			FROM transactions
			WHERE user_id = $1
			  AND type = 'EXPENSE'
			  AND transaction_date >= $2
			  AND transaction_date <= $3
			  AND deleted_at IS NULL
			GROUP BY category_id
		) t ON t.category_id = b.category_id
		WHERE b.user_id = $1
		  AND b.cycle_start = $2
		  AND b.cycle_end = $3
		  AND b.deleted_at IS NULL
		ORDER BY c.name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID, cycleStart, cycleEnd)
	if err != nil {
		return nil, fmt.Errorf("get budget progress: %w", err)
	}
	defer rows.Close()

	var progresses []BudgetProgress
	for rows.Next() {
		var bp BudgetProgress
		err := rows.Scan(
			&bp.CategoryID,
			&bp.CategoryName,
			&bp.CategoryIcon,
			&bp.CategoryColor,
			&bp.Planned,
			&bp.Actual,
		)
		if err != nil {
			return nil, fmt.Errorf("scan budget progress: %w", err)
		}
		bp.Variance = bp.Planned - bp.Actual
		bp.Overspent = bp.Variance < 0
		progresses = append(progresses, bp)
	}

	if progresses == nil {
		progresses = []BudgetProgress{}
	}

	return progresses, nil
}

// GetAccountSummaries retrieves all active accounts with current balances and computes total assets.
func (r *PostgresRepository) GetAccountSummaries(ctx context.Context, userID uuid.UUID) ([]AccountSummary, int64, error) {
	query := `
		SELECT id, name, type, current_balance
		FROM accounts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("get account summaries: %w", err)
	}
	defer rows.Close()

	var accounts []AccountSummary
	var totalAssets int64
	for rows.Next() {
		var acc AccountSummary
		if err := rows.Scan(&acc.ID, &acc.Name, &acc.Type, &acc.Balance); err != nil {
			return nil, 0, fmt.Errorf("scan account summary: %w", err)
		}
		totalAssets += acc.Balance
		accounts = append(accounts, acc)
	}

	if accounts == nil {
		accounts = []AccountSummary{}
	}

	return accounts, totalAssets, nil
}

// GetTotalAssetsAsOf calculates historical total assets as of a given date (sum of opening balances + net income up to asOfDate).
func (r *PostgresRepository) GetTotalAssetsAsOf(ctx context.Context, userID uuid.UUID, asOfDate string) (int64, error) {
	query := `
		WITH opening AS (
			SELECT COALESCE(SUM(opening_balance), 0) AS total_opening
			FROM accounts
			WHERE user_id = $1 AND deleted_at IS NULL
		),
		flows AS (
			SELECT
				COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) -
				COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS net_flow
			FROM transactions
			WHERE user_id = $1
			  AND transaction_date <= $2
			  AND deleted_at IS NULL
		)
		SELECT opening.total_opening + flows.net_flow
		FROM opening, flows;
	`

	var totalAssets int64
	err := r.pool.QueryRow(ctx, query, userID, asOfDate).Scan(&totalAssets)
	if err != nil {
		return 0, fmt.Errorf("get total assets as of %s: %w", asOfDate, err)
	}

	return totalAssets, nil
}

// GetReceivablesAndInvestmentsTotals computes active unpaid receivables and total investment valuation.
func (r *PostgresRepository) GetReceivablesAndInvestmentsTotals(ctx context.Context, userID uuid.UUID) (int64, int64, error) {
	recQuery := `
		SELECT COALESCE(SUM(r.principal - COALESCE(p.paid, 0)), 0)
		FROM receivables r
		LEFT JOIN (
			SELECT receivable_id, SUM(amount) AS paid
			FROM receivable_payments
			GROUP BY receivable_id
		) p ON p.receivable_id = r.id
		WHERE r.user_id = $1
		  AND r.deleted_at IS NULL
		  AND r.status NOT IN ('PAID', 'WRITTEN_OFF');
	`
	var totalReceivables int64
	if err := r.pool.QueryRow(ctx, recQuery, userID).Scan(&totalReceivables); err != nil {
		return 0, 0, fmt.Errorf("get total receivables: %w", err)
	}
	if totalReceivables < 0 {
		totalReceivables = 0
	}

	invQuery := `
		SELECT COALESCE(SUM(current_value), 0)
		FROM investments
		WHERE user_id = $1 AND deleted_at IS NULL;
	`
	var totalInvestments int64
	if err := r.pool.QueryRow(ctx, invQuery, userID).Scan(&totalInvestments); err != nil {
		return 0, 0, fmt.Errorf("get total investments: %w", err)
	}

	return totalReceivables, totalInvestments, nil
}
