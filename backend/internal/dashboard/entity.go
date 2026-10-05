package dashboard

import (
	"github.com/google/uuid"
)

// CycleSummary represents the financial cycle window and progress.
type CycleSummary struct {
	StartDate     string `json:"start_date"`     // YYYY-MM-DD
	EndDate       string `json:"end_date"`       // YYYY-MM-DD
	CycleStartDay int    `json:"cycle_start_day"`
	DayOfCycle    int    `json:"day_of_cycle"`
	TotalDays     int    `json:"total_days"`
	DaysRemaining int    `json:"days_remaining"`
}

// BudgetProgress represents category budget usage within the cycle.
type BudgetProgress struct {
	CategoryID    uuid.UUID `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	CategoryIcon  string    `json:"category_icon"`
	CategoryColor string    `json:"category_color"`
	Planned       int64     `json:"planned"`
	Actual        int64     `json:"actual"`
	Variance      int64     `json:"variance"` // planned - actual
	Overspent     bool      `json:"overspent"`
}

// AccountSummary represents an individual account and its balance.
type AccountSummary struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Balance int64     `json:"balance"`
}

// DashboardData is the comprehensive aggregation response for the dashboard view.
type DashboardData struct {
	Cycle               CycleSummary     `json:"cycle"`
	Income              int64            `json:"income"`
	Expense             int64            `json:"expense"`
	NetCashFlow         int64            `json:"net_cash_flow"`
	TotalAssets         int64            `json:"total_assets"`
	TotalReceivables    int64            `json:"total_receivables"`
	TotalInvestments    int64            `json:"total_investments"`
	PreviousCycleAssets int64            `json:"previous_cycle_assets"`
	AssetGrowth         int64            `json:"asset_growth"`
	Budgets             []BudgetProgress `json:"budgets"`
	Accounts            []AccountSummary `json:"accounts"`
}
