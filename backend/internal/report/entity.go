package report

import (
	"errors"

	"github.com/google/uuid"
)

var (
	// ErrInvalidYear indicates year is outside a sensible range.
	ErrInvalidYear = errors.New("invalid year: must be between 2000 and 2100")
	// ErrInvalidMonth indicates month is outside 1 to 12.
	ErrInvalidMonth = errors.New("invalid month: must be between 1 and 12")
	// ErrInvalidDateFormat indicates a date cannot be parsed as YYYY-MM-DD.
	ErrInvalidDateFormat = errors.New("invalid date format: expected YYYY-MM-DD")
)

// ReportData represents the financial report response for a period (monthly or cycle).
type ReportData struct {
	Period               string               `json:"period"`
	StartDate            string               `json:"start_date"`
	EndDate              string               `json:"end_date"`
	Income               int64                `json:"income"`
	FormattedIncome      string               `json:"formatted_income,omitempty"`
	Expense              int64                `json:"expense"`
	FormattedExpense     string               `json:"formatted_expense,omitempty"`
	Transfer             int64                `json:"transfer"`
	FormattedTransfer    string               `json:"formatted_transfer,omitempty"`
	NetCashFlow          int64                `json:"net_cash_flow"`
	FormattedNetCashFlow string               `json:"formatted_net_cash_flow,omitempty"`
	BudgetVsActual       []BudgetVsActualItem `json:"budget_vs_actual"`
	Assets               AssetSnapshot        `json:"assets"`
	Liabilities          int64                `json:"liabilities"`
	FormattedLiabilities string               `json:"formatted_liabilities,omitempty"`
	NetWorth             int64                `json:"net_worth"`
	FormattedNetWorth    string               `json:"formatted_net_worth,omitempty"`
}

// BudgetVsActualItem represents planned vs actual expense comparison for a category.
type BudgetVsActualItem struct {
	CategoryID        uuid.UUID `json:"category_id,omitempty"`
	CategoryName      string    `json:"category"`
	CategoryIcon      string    `json:"category_icon,omitempty"`
	Budget            int64     `json:"budget"`
	FormattedBudget   string    `json:"formatted_budget,omitempty"`
	Actual            int64     `json:"actual"`
	FormattedActual   string    `json:"formatted_actual,omitempty"`
	Variance          int64     `json:"variance"`
	FormattedVariance string    `json:"formatted_variance,omitempty"`
}

// AssetSnapshot represents the breakdown of all asset types.
type AssetSnapshot struct {
	Accounts             int64                 `json:"accounts"`
	FormattedAccounts    string                `json:"formatted_accounts,omitempty"`
	Receivables          int64                 `json:"receivables"`
	FormattedReceivables string                `json:"formatted_receivables,omitempty"`
	Investments          int64                 `json:"investments"`
	FormattedInvestments string                `json:"formatted_investments,omitempty"`
	Total                int64                 `json:"total"`
	FormattedTotal       string                `json:"formatted_total,omitempty"`
	AccountList          []AccountAssetItem    `json:"account_list,omitempty"`
	ReceivableList       []ReceivableAssetItem `json:"receivable_list,omitempty"`
	InvestmentList       []InvestmentAssetItem `json:"investment_list,omitempty"`
}

// AccountAssetItem represents an individual cash account asset.
type AccountAssetItem struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Balance int64     `json:"balance"`
}

// ReceivableAssetItem represents an individual receivable asset.
type ReceivableAssetItem struct {
	ID           uuid.UUID `json:"id"`
	Counterparty string    `json:"counterparty"`
	Principal    int64     `json:"principal"`
	Remaining    int64     `json:"remaining_amount"`
	Status       string    `json:"status"`
}

// InvestmentAssetItem represents an individual investment asset.
type InvestmentAssetItem struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Capital      int64     `json:"capital"`
	CurrentValue int64     `json:"current_value"`
}
