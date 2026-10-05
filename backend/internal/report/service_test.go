package report_test

import (
	"context"
	"testing"

	"github.com/fernando/financial-management-tracking/backend/internal/report"
	"github.com/google/uuid"
)

type mockRepo struct {
	startDay       int
	startDayErr    error
	income         int64
	expense        int64
	transfer       int64
	cashFlowErr    error
	budgetVsActual []report.BudgetVsActualItem
	bvaErr         error
	assetSnapshot  report.AssetSnapshot
	assetErr       error

	lastStartDate string
	lastEndDate   string
}

func (m *mockRepo) GetUserCycleStartDay(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.startDayErr != nil {
		return 0, m.startDayErr
	}
	if m.startDay == 0 {
		return 1, nil
	}
	return m.startDay, nil
}

func (m *mockRepo) GetCashFlow(ctx context.Context, userID uuid.UUID, startDate, endDate string) (int64, int64, int64, error) {
	m.lastStartDate = startDate
	m.lastEndDate = endDate
	if m.cashFlowErr != nil {
		return 0, 0, 0, m.cashFlowErr
	}
	return m.income, m.expense, m.transfer, nil
}

func (m *mockRepo) GetBudgetVsActual(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]report.BudgetVsActualItem, error) {
	if m.bvaErr != nil {
		return nil, m.bvaErr
	}
	return m.budgetVsActual, nil
}

func (m *mockRepo) GetAssetSnapshot(ctx context.Context, userID uuid.UUID) (report.AssetSnapshot, error) {
	if m.assetErr != nil {
		return report.AssetSnapshot{}, m.assetErr
	}
	return m.assetSnapshot, nil
}

func TestGenerateMonthlyReport_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mock := &mockRepo{
		income:   13000000,
		expense:  4300000,
		transfer: 1200000,
		budgetVsActual: []report.BudgetVsActualItem{
			{
				CategoryName: "Operational",
				Budget:       4000000,
				Actual:       4300000,
				Variance:     -300000,
			},
			{
				CategoryName: "Savings",
				Budget:       5000000,
				Actual:       5000000,
				Variance:     0,
			},
		},
		assetSnapshot: report.AssetSnapshot{
			Accounts:    9000000,
			Receivables: 2000000,
			Investments: 2250000,
			Total:       13250000,
		},
	}

	svc := report.NewService(mock)
	res, err := svc.GenerateMonthlyReport(ctx, userID, 2026, 9)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.Period != "2026-09" {
		t.Errorf("expected period '2026-09', got %s", res.Period)
	}
	if res.StartDate != "2026-09-01" {
		t.Errorf("expected start date '2026-09-01', got %s", res.StartDate)
	}
	if res.EndDate != "2026-09-30" {
		t.Errorf("expected end date '2026-09-30', got %s", res.EndDate)
	}
	if res.Income != 13000000 {
		t.Errorf("expected income 13000000, got %d", res.Income)
	}
	if res.Expense != 4300000 {
		t.Errorf("expected expense 4300000, got %d", res.Expense)
	}
	if res.Transfer != 1200000 {
		t.Errorf("expected transfer 1200000, got %d", res.Transfer)
	}
	if res.NetCashFlow != 8700000 {
		t.Errorf("expected net cash flow 8700000, got %d", res.NetCashFlow)
	}
	if len(res.BudgetVsActual) != 2 {
		t.Fatalf("expected 2 budget vs actual items, got %d", len(res.BudgetVsActual))
	}
	if res.Assets.Total != 13250000 {
		t.Errorf("expected total assets 13250000, got %d", res.Assets.Total)
	}
	if res.NetWorth != 13250000 {
		t.Errorf("expected net worth 13250000, got %d", res.NetWorth)
	}
}

func TestGenerateMonthlyReport_Validation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	mock := &mockRepo{}
	svc := report.NewService(mock)

	// Invalid year low
	_, err := svc.GenerateMonthlyReport(ctx, userID, 1999, 5)
	if err != report.ErrInvalidYear {
		t.Errorf("expected ErrInvalidYear, got %v", err)
	}

	// Invalid year high
	_, err = svc.GenerateMonthlyReport(ctx, userID, 2101, 5)
	if err != report.ErrInvalidYear {
		t.Errorf("expected ErrInvalidYear, got %v", err)
	}

	// Invalid month low
	_, err = svc.GenerateMonthlyReport(ctx, userID, 2026, 0)
	if err != report.ErrInvalidMonth {
		t.Errorf("expected ErrInvalidMonth, got %v", err)
	}

	// Invalid month high
	_, err = svc.GenerateMonthlyReport(ctx, userID, 2026, 13)
	if err != report.ErrInvalidMonth {
		t.Errorf("expected ErrInvalidMonth, got %v", err)
	}
}

func TestGenerateMonthlyReport_NoTransactions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	mock := &mockRepo{
		income:         0,
		expense:        0,
		transfer:       0,
		budgetVsActual: []report.BudgetVsActualItem{},
		assetSnapshot: report.AssetSnapshot{
			Accounts:    0,
			Receivables: 0,
			Investments: 0,
			Total:       0,
		},
	}

	svc := report.NewService(mock)
	res, err := svc.GenerateMonthlyReport(ctx, userID, 2026, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.Period != "2026-02" {
		t.Errorf("expected period '2026-02', got %s", res.Period)
	}
	if res.StartDate != "2026-02-01" {
		t.Errorf("expected start date '2026-02-01', got %s", res.StartDate)
	}
	// Non-leap year 2026 -> 28 days
	if res.EndDate != "2026-02-28" {
		t.Errorf("expected end date '2026-02-28', got %s", res.EndDate)
	}
	if res.Income != 0 || res.Expense != 0 || res.NetCashFlow != 0 {
		t.Errorf("expected zero cash flow, got %+v", res)
	}
}

func TestGenerateCycleReport_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mock := &mockRepo{
		startDay: 25,
		income:   20000000,
		expense:  10000000,
		transfer: 2500000,
		budgetVsActual: []report.BudgetVsActualItem{
			{
				CategoryName: "Living",
				Budget:       10000000,
				Actual:       9500000,
				Variance:     500000,
			},
		},
		assetSnapshot: report.AssetSnapshot{
			Accounts:    15000000,
			Receivables: 5000000,
			Investments: 10000000,
			Total:       30000000,
		},
	}

	svc := report.NewService(mock)
	res, err := svc.GenerateCycleReport(ctx, userID, "2026-09-10")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// For startDay 25 and date 2026-09-10, cycle is 2026-08-25 to 2026-09-24
	if res.StartDate != "2026-08-25" {
		t.Errorf("expected start date 2026-08-25, got %s", res.StartDate)
	}
	if res.EndDate != "2026-09-24" {
		t.Errorf("expected end date 2026-09-24, got %s", res.EndDate)
	}
	if res.Period != "2026-08-25 to 2026-09-24" {
		t.Errorf("expected period '2026-08-25 to 2026-09-24', got %s", res.Period)
	}
	if res.Income != 20000000 {
		t.Errorf("expected income 20000000, got %d", res.Income)
	}
	if res.NetCashFlow != 10000000 {
		t.Errorf("expected net cash flow 10000000, got %d", res.NetCashFlow)
	}
	if res.NetWorth != 30000000 {
		t.Errorf("expected net worth 30000000, got %d", res.NetWorth)
	}
}

func TestGenerateCycleReport_InvalidDate(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	mock := &mockRepo{}
	svc := report.NewService(mock)

	_, err := svc.GenerateCycleReport(ctx, userID, "invalid-date")
	if err != report.ErrInvalidDateFormat {
		t.Errorf("expected ErrInvalidDateFormat, got %v", err)
	}
}
