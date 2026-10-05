package cycle

import (
	"errors"
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format: must be YYYY-MM-DD")
	ErrSettingsNotFound  = errors.New("user settings not found")
)

// CycleInfo encapsulates the calculated financial cycle boundaries and progress indicators.
type CycleInfo struct {
	CycleStartDay int    `json:"cycle_start_day"`
	StartDate     string `json:"start_date"`     // YYYY-MM-DD
	EndDate       string `json:"end_date"`       // YYYY-MM-DD
	TotalDays     int    `json:"total_days"`     // Total days in this cycle
	DayOfCycle    int    `json:"day_of_cycle"`    // 1-indexed current day within cycle
	DaysRemaining int    `json:"days_remaining"` // Days left until cycle end
}

// CycleSummary aggregates financial figures within the cycle boundary.
type CycleSummary struct {
	Cycle            CycleInfo `json:"cycle"`
	TotalIncome      int64     `json:"total_income"`  // Sum of INCOME transactions (IDR)
	TotalExpense     int64     `json:"total_expense"` // Sum of EXPENSE transactions (IDR)
	NetSavings       int64     `json:"net_savings"`   // TotalIncome - TotalExpense (IDR)
	TransactionCount int       `json:"transaction_count"`
}
