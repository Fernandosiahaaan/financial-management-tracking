package dashboard

import (
	"context"
	"time"

	"github.com/fernando/financial-management-tracking/backend/internal/cycle"
	"github.com/google/uuid"
)

// Service defines business operations for the dashboard.
type Service interface {
	GetDashboardData(ctx context.Context, userID uuid.UUID, targetDateStr string) (*DashboardData, error)
}

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// NewService creates a new dashboard Service.
func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

// GetDashboardData aggregates comprehensive cycle, cash flow, budget, and asset metrics.
func (s *DefaultService) GetDashboardData(ctx context.Context, userID uuid.UUID, targetDateStr string) (*DashboardData, error) {
	var targetDate time.Time
	if targetDateStr == "" {
		targetDate = time.Now().UTC()
	} else {
		parsed, err := time.Parse("2006-01-02", targetDateStr)
		if err != nil {
			return nil, cycle.ErrInvalidDateFormat
		}
		targetDate = parsed
	}

	startDay, err := s.repo.GetUserCycleStartDay(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Calculate current cycle
	cycleStart, cycleEnd := cycle.CalculateCycle(startDay, targetDate)
	startDateStr := cycleStart.Format("2006-01-02")
	endDateStr := cycleEnd.Format("2006-01-02")

	totalDays := int(cycleEnd.Sub(cycleStart).Hours()/24) + 1
	dayOfCycle := int(targetDate.Sub(cycleStart).Hours()/24) + 1
	if dayOfCycle < 1 {
		dayOfCycle = 1
	} else if dayOfCycle > totalDays {
		dayOfCycle = totalDays
	}
	daysRemaining := totalDays - dayOfCycle

	// 2. Calculate previous cycle (evaluating the day before current cycle start)
	_, prevEnd := cycle.CalculateCycle(startDay, cycleStart.AddDate(0, 0, -1))
	prevEndDateStr := prevEnd.Format("2006-01-02")

	// 3. Aggregate cycle cash flow
	income, expense, err := s.repo.GetCycleCashFlow(ctx, userID, startDateStr, endDateStr)
	if err != nil {
		return nil, err
	}
	netCashFlow := income - expense

	// 4. Retrieve budget progress with live variance
	budgets, err := s.repo.GetBudgetProgress(ctx, userID, startDateStr, endDateStr)
	if err != nil {
		return nil, err
	}

	// 5. Retrieve active accounts and cash balance
	accounts, totalCash, err := s.repo.GetAccountSummaries(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 5b. Retrieve active receivables and investments totals
	totalReceivables, totalInvestments, err := s.repo.GetReceivablesAndInvestmentsTotals(ctx, userID)
	if err != nil {
		return nil, err
	}
	totalAssets := totalCash + totalReceivables + totalInvestments

	// 6. Calculate total assets as of the end of the previous cycle
	prevAssets, err := s.repo.GetTotalAssetsAsOf(ctx, userID, prevEndDateStr)
	if err != nil {
		return nil, err
	}
	assetGrowth := totalAssets - prevAssets

	return &DashboardData{
		Cycle: CycleSummary{
			StartDate:     startDateStr,
			EndDate:       endDateStr,
			CycleStartDay: startDay,
			DayOfCycle:    dayOfCycle,
			TotalDays:     totalDays,
			DaysRemaining: daysRemaining,
		},
		Income:              income,
		Expense:             expense,
		NetCashFlow:         netCashFlow,
		TotalAssets:         totalAssets,
		TotalReceivables:    totalReceivables,
		TotalInvestments:    totalInvestments,
		PreviousCycleAssets: prevAssets,
		AssetGrowth:         assetGrowth,
		Budgets:             budgets,
		Accounts:            accounts,
	}, nil
}
