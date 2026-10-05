package dashboard

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockRepository struct {
	startDay         int
	income           int64
	expense          int64
	budgets          []BudgetProgress
	accounts         []AccountSummary
	totalAssets      int64
	totalReceivables int64
	totalInvestments int64
	prevAssets       int64
	cashFlowErr      error
	budgetErr        error
	accountErr       error
	prevAssetErr     error
	userStartDayErr  error
}

func (m *mockRepository) GetUserCycleStartDay(_ context.Context, _ uuid.UUID) (int, error) {
	if m.userStartDayErr != nil {
		return 0, m.userStartDayErr
	}
	return m.startDay, nil
}

func (m *mockRepository) GetCycleCashFlow(_ context.Context, _ uuid.UUID, _, _ string) (int64, int64, error) {
	if m.cashFlowErr != nil {
		return 0, 0, m.cashFlowErr
	}
	return m.income, m.expense, nil
}

func (m *mockRepository) GetBudgetProgress(_ context.Context, _ uuid.UUID, _, _ string) ([]BudgetProgress, error) {
	if m.budgetErr != nil {
		return nil, m.budgetErr
	}
	return m.budgets, nil
}

func (m *mockRepository) GetAccountSummaries(_ context.Context, _ uuid.UUID) ([]AccountSummary, int64, error) {
	if m.accountErr != nil {
		return nil, 0, m.accountErr
	}
	return m.accounts, m.totalAssets, nil
}

func (m *mockRepository) GetReceivablesAndInvestmentsTotals(_ context.Context, _ uuid.UUID) (int64, int64, error) {
	return m.totalReceivables, m.totalInvestments, nil
}

func (m *mockRepository) GetTotalAssetsAsOf(_ context.Context, _ uuid.UUID, _ string) (int64, error) {
	if m.prevAssetErr != nil {
		return 0, m.prevAssetErr
	}
	return m.prevAssets, nil
}

func TestGetDashboardData_Success(t *testing.T) {
	repo := &mockRepository{
		startDay: 25,
		income:   10000000,
		expense:  4000000,
		budgets: []BudgetProgress{
			{
				CategoryID:   uuid.New(),
				CategoryName: "Groceries",
				Planned:      5000000,
				Actual:       4000000,
				Variance:     1000000,
				Overspent:    false,
			},
		},
		accounts: []AccountSummary{
			{ID: uuid.New(), Name: "Checking", Type: "BANK", Balance: 15000000},
		},
		totalAssets: 15000000,
		prevAssets:  9000000,
	}

	svc := NewService(repo)

	data, err := svc.GetDashboardData(context.Background(), uuid.New(), "2026-08-30")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if data.Cycle.StartDate != "2026-08-25" || data.Cycle.EndDate != "2026-09-24" {
		t.Errorf("unexpected cycle dates: %s to %s", data.Cycle.StartDate, data.Cycle.EndDate)
	}
	if data.Income != 10000000 {
		t.Errorf("expected income 10000000, got %d", data.Income)
	}
	if data.Expense != 4000000 {
		t.Errorf("expected expense 4000000, got %d", data.Expense)
	}
	if data.NetCashFlow != 6000000 {
		t.Errorf("expected net cash flow 6000000, got %d", data.NetCashFlow)
	}
	if data.TotalAssets != 15000000 {
		t.Errorf("expected total assets 15000000, got %d", data.TotalAssets)
	}
	if data.PreviousCycleAssets != 9000000 {
		t.Errorf("expected previous cycle assets 9000000, got %d", data.PreviousCycleAssets)
	}
	if data.AssetGrowth != 6000000 {
		t.Errorf("expected asset growth 6000000, got %d", data.AssetGrowth)
	}
	if len(data.Budgets) != 1 || len(data.Accounts) != 1 {
		t.Errorf("unexpected budgets or accounts length: %d budgets, %d accounts", len(data.Budgets), len(data.Accounts))
	}
}

func TestGetDashboardData_InvalidDate(t *testing.T) {
	repo := &mockRepository{startDay: 1}
	svc := NewService(repo)

	_, err := svc.GetDashboardData(context.Background(), uuid.New(), "invalid-date")
	if err == nil {
		t.Errorf("expected error for invalid date, got nil")
	}
}

func TestGetDashboardData_ZeroMetrics(t *testing.T) {
	repo := &mockRepository{
		startDay:    1,
		income:      0,
		expense:     0,
		budgets:     []BudgetProgress{},
		accounts:    []AccountSummary{},
		totalAssets: 0,
		prevAssets:  0,
	}
	svc := NewService(repo)

	data, err := svc.GetDashboardData(context.Background(), uuid.New(), "2026-10-01")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.NetCashFlow != 0 {
		t.Errorf("expected 0 net cash flow, got %d", data.NetCashFlow)
	}
	if data.AssetGrowth != 0 {
		t.Errorf("expected 0 asset growth, got %d", data.AssetGrowth)
	}
}

func TestGetDashboardData_NegativeGrowth(t *testing.T) {
	repo := &mockRepository{
		startDay:    1,
		income:      2000000,
		expense:     5000000,
		totalAssets: 7000000,
		prevAssets:  10000000,
	}
	svc := NewService(repo)

	data, err := svc.GetDashboardData(context.Background(), uuid.New(), "2026-10-01")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.AssetGrowth != -3000000 {
		t.Errorf("expected -3000000 asset growth, got %d", data.AssetGrowth)
	}
	if data.NetCashFlow != -3000000 {
		t.Errorf("expected -3000000 net cash flow, got %d", data.NetCashFlow)
	}
}

func TestGetDashboardData_WithReceivablesAndInvestments(t *testing.T) {
	repo := &mockRepository{
		startDay:         1,
		income:           10000000,
		expense:          4000000,
		totalAssets:      20000000, // cash
		totalReceivables: 5000000,  // receivables
		totalInvestments: 15000000, // investments
		prevAssets:       30000000,
	}
	svc := NewService(repo)

	data, err := svc.GetDashboardData(context.Background(), uuid.New(), "2026-10-15")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.TotalAssets != 40000000 {
		t.Errorf("expected total assets 40000000 (20M cash + 5M rec + 15M inv), got %d", data.TotalAssets)
	}
	if data.TotalReceivables != 5000000 {
		t.Errorf("expected total receivables 5000000, got %d", data.TotalReceivables)
	}
	if data.TotalInvestments != 15000000 {
		t.Errorf("expected total investments 15000000, got %d", data.TotalInvestments)
	}
	if data.AssetGrowth != 10000000 {
		t.Errorf("expected asset growth 10000000 (40M - 30M), got %d", data.AssetGrowth)
	}
}

