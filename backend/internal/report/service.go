package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/cycle"
	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Service defines reporting business logic.
type Service interface {
	GenerateMonthlyReport(ctx context.Context, userID uuid.UUID, year, month int) (*ReportData, error)
	GenerateCycleReport(ctx context.Context, userID uuid.UUID, targetDateStr string) (*ReportData, error)
}

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// NewService creates a new report Service.
func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

// GenerateMonthlyReport generates a calendar-month financial report.
func (s *DefaultService) GenerateMonthlyReport(ctx context.Context, userID uuid.UUID, year, month int) (*ReportData, error) {
	if year < 2000 || year > 2100 {
		return nil, ErrInvalidYear
	}
	if month < 1 || month > 12 {
		return nil, ErrInvalidMonth
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")
	period := fmt.Sprintf("%04d-%02d", year, month)

	return s.buildReport(ctx, userID, period, startDateStr, endDateStr)
}

// GenerateCycleReport generates a financial-cycle financial report based on user cycle settings and a reference date.
func (s *DefaultService) GenerateCycleReport(ctx context.Context, userID uuid.UUID, targetDateStr string) (*ReportData, error) {
	targetDate, err := parseDate(targetDateStr)
	if err != nil {
		return nil, err
	}

	startDay, err := s.repo.GetUserCycleStartDay(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup user cycle start day: %w", err)
	}

	cycleStart, cycleEnd := cycle.CalculateCycle(startDay, targetDate)
	startDateStr := cycleStart.Format("2006-01-02")
	endDateStr := cycleEnd.Format("2006-01-02")
	period := fmt.Sprintf("%s to %s", startDateStr, endDateStr)

	return s.buildReport(ctx, userID, period, startDateStr, endDateStr)
}

func (s *DefaultService) buildReport(ctx context.Context, userID uuid.UUID, period, startDateStr, endDateStr string) (*ReportData, error) {
	// 1. Cash flow
	income, expense, transfer, err := s.repo.GetCashFlow(ctx, userID, startDateStr, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("generate cash flow: %w", err)
	}

	// 2. Budget vs actual
	budgetVsActual, err := s.repo.GetBudgetVsActual(ctx, userID, startDateStr, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("generate budget vs actual: %w", err)
	}

	// 3. Asset snapshot
	assets, err := s.repo.GetAssetSnapshot(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("generate asset snapshot: %w", err)
	}

	liabilities := int64(0)
	netWorth := assets.Total - liabilities
	netCashFlow := income - expense

	for i := range budgetVsActual {
		budgetVsActual[i].FormattedBudget = money.FormatIDR(budgetVsActual[i].Budget)
		budgetVsActual[i].FormattedActual = money.FormatIDR(budgetVsActual[i].Actual)
		budgetVsActual[i].FormattedVariance = money.FormatIDR(budgetVsActual[i].Variance)
	}

	assets.FormattedAccounts = money.FormatIDR(assets.Accounts)
	assets.FormattedReceivables = money.FormatIDR(assets.Receivables)
	assets.FormattedInvestments = money.FormatIDR(assets.Investments)
	assets.FormattedTotal = money.FormatIDR(assets.Total)

	return &ReportData{
		Period:               period,
		StartDate:            startDateStr,
		EndDate:              endDateStr,
		Income:               income,
		FormattedIncome:      money.FormatIDR(income),
		Expense:              expense,
		FormattedExpense:     money.FormatIDR(expense),
		Transfer:             transfer,
		FormattedTransfer:    money.FormatIDR(transfer),
		NetCashFlow:          netCashFlow,
		FormattedNetCashFlow: money.FormatIDR(netCashFlow),
		BudgetVsActual:       budgetVsActual,
		Assets:               assets,
		Liabilities:          liabilities,
		FormattedLiabilities: money.FormatIDR(liabilities),
		NetWorth:             netWorth,
		FormattedNetWorth:    money.FormatIDR(netWorth),
	}, nil
}

func parseDate(str string) (time.Time, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}
