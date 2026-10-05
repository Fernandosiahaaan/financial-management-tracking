package report

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access operations for reporting and asset snapshots.
type Repository interface {
	GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error)
	GetCashFlow(ctx context.Context, userID uuid.UUID, startDate, endDate string) (income, expense, transfer int64, err error)
	GetBudgetVsActual(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]BudgetVsActualItem, error)
	GetAssetSnapshot(ctx context.Context, userID uuid.UUID) (AssetSnapshot, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository for reports.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// GetUserCycleStartDay retrieves the user's cycle start day, defaulting to 1 if not set.
func (r *PostgresRepository) GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT cycle_start_day FROM user_settings WHERE user_id = $1;`
	var startDay int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&startDay)
	if err != nil {
		return 1, nil
	}
	return startDay, nil
}

// GetCashFlow calculates total income, expense, and transfer volume in the given date window.
func (r *PostgresRepository) GetCashFlow(ctx context.Context, userID uuid.UUID, startDate, endDate string) (int64, int64, int64, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS total_expense,
			COALESCE(SUM(CASE WHEN type = 'TRANSFER' THEN amount ELSE 0 END), 0) AS total_transfer
		FROM transactions
		WHERE user_id = $1
		  AND transaction_date >= $2
		  AND transaction_date <= $3
		  AND deleted_at IS NULL;
	`

	var income, expense, transfer int64
	err := r.pool.QueryRow(ctx, query, userID, startDate, endDate).Scan(&income, &expense, &transfer)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("get cash flow: %w", err)
	}

	return income, expense, transfer, nil
}

// GetBudgetVsActual compares planned budget vs actual spending per category for the date window.
func (r *PostgresRepository) GetBudgetVsActual(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]BudgetVsActualItem, error) {
	query := `
		WITH period_budgets AS (
			SELECT category_id, SUM(planned_amount) AS planned_amount
			FROM budgets
			WHERE user_id = $1
			  AND deleted_at IS NULL
			  AND (
				  (cycle_start = $2 AND cycle_end = $3)
				  OR
				  (NOT EXISTS (
					  SELECT 1 FROM budgets b2
					  WHERE b2.user_id = $1
					    AND b2.cycle_start = $2
					    AND b2.cycle_end = $3
					    AND b2.deleted_at IS NULL
				  ) AND cycle_start <= $3 AND cycle_end >= $2)
			  )
			GROUP BY category_id
		),
		period_expenses AS (
			SELECT category_id, SUM(amount) AS actual_amount
			FROM transactions
			WHERE user_id = $1
			  AND type = 'EXPENSE'
			  AND transaction_date >= $2
			  AND transaction_date <= $3
			  AND deleted_at IS NULL
			GROUP BY category_id
		),
		combined AS (
			SELECT
				COALESCE(b.category_id, e.category_id) AS category_id,
				COALESCE(b.planned_amount, 0) AS budget,
				COALESCE(e.actual_amount, 0) AS actual
			FROM period_budgets b
			FULL OUTER JOIN period_expenses e ON b.category_id = e.category_id
		)
		SELECT
			c.id,
			c.name,
			COALESCE(c.icon, ''),
			cb.budget,
			cb.actual
		FROM combined cb
		JOIN categories c ON c.id = cb.category_id AND c.deleted_at IS NULL
		ORDER BY c.name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get budget vs actual: %w", err)
	}
	defer rows.Close()

	var items []BudgetVsActualItem
	for rows.Next() {
		var item BudgetVsActualItem
		if err := rows.Scan(
			&item.CategoryID,
			&item.CategoryName,
			&item.CategoryIcon,
			&item.Budget,
			&item.Actual,
		); err != nil {
			return nil, fmt.Errorf("scan budget vs actual item: %w", err)
		}
		item.Variance = item.Budget - item.Actual
		items = append(items, item)
	}

	if items == nil {
		items = []BudgetVsActualItem{}
	}

	return items, nil
}

// GetAssetSnapshot gathers all current accounts, active receivables, and investments.
func (r *PostgresRepository) GetAssetSnapshot(ctx context.Context, userID uuid.UUID) (AssetSnapshot, error) {
	snapshot := AssetSnapshot{
		AccountList:    []AccountAssetItem{},
		ReceivableList: []ReceivableAssetItem{},
		InvestmentList: []InvestmentAssetItem{},
	}

	// 1. Accounts
	accQuery := `
		SELECT id, name, type, current_balance
		FROM accounts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC;
	`
	accRows, err := r.pool.Query(ctx, accQuery, userID)
	if err != nil {
		return snapshot, fmt.Errorf("get accounts snapshot: %w", err)
	}
	defer accRows.Close()

	for accRows.Next() {
		var acc AccountAssetItem
		if err := accRows.Scan(&acc.ID, &acc.Name, &acc.Type, &acc.Balance); err != nil {
			return snapshot, fmt.Errorf("scan account asset: %w", err)
		}
		snapshot.Accounts += acc.Balance
		snapshot.AccountList = append(snapshot.AccountList, acc)
	}

	// 2. Active Receivables
	recQuery := `
		SELECT
			r.id,
			r.counterparty,
			r.principal,
			(r.principal - COALESCE(p.paid, 0)) AS remaining,
			r.status
		FROM receivables r
		LEFT JOIN (
			SELECT receivable_id, SUM(amount) AS paid
			FROM receivable_payments
			GROUP BY receivable_id
		) p ON p.receivable_id = r.id
		WHERE r.user_id = $1
		  AND r.deleted_at IS NULL
		  AND r.status NOT IN ('PAID', 'WRITTEN_OFF')
		ORDER BY r.created_at DESC;
	`
	recRows, err := r.pool.Query(ctx, recQuery, userID)
	if err != nil {
		return snapshot, fmt.Errorf("get receivables snapshot: %w", err)
	}
	defer recRows.Close()

	for recRows.Next() {
		var rec ReceivableAssetItem
		if err := recRows.Scan(&rec.ID, &rec.Counterparty, &rec.Principal, &rec.Remaining, &rec.Status); err != nil {
			return snapshot, fmt.Errorf("scan receivable asset: %w", err)
		}
		if rec.Remaining < 0 {
			rec.Remaining = 0
		}
		snapshot.Receivables += rec.Remaining
		snapshot.ReceivableList = append(snapshot.ReceivableList, rec)
	}

	// 3. Investments
	invQuery := `
		SELECT id, name, type, capital, current_value
		FROM investments
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC;
	`
	invRows, err := r.pool.Query(ctx, invQuery, userID)
	if err != nil {
		return snapshot, fmt.Errorf("get investments snapshot: %w", err)
	}
	defer invRows.Close()

	for invRows.Next() {
		var inv InvestmentAssetItem
		if err := invRows.Scan(&inv.ID, &inv.Name, &inv.Type, &inv.Capital, &inv.CurrentValue); err != nil {
			return snapshot, fmt.Errorf("scan investment asset: %w", err)
		}
		snapshot.Investments += inv.CurrentValue
		snapshot.InvestmentList = append(snapshot.InvestmentList, inv)
	}

	snapshot.Total = snapshot.Accounts + snapshot.Receivables + snapshot.Investments
	return snapshot, nil
}
