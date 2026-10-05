package report_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fernando/financial-management-tracking/backend/internal/account"
	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/fernando/financial-management-tracking/backend/internal/budget"
	"github.com/fernando/financial-management-tracking/backend/internal/category"
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/fernando/financial-management-tracking/backend/internal/investment"
	"github.com/fernando/financial-management-tracking/backend/internal/receivable"
	"github.com/fernando/financial-management-tracking/backend/internal/report"
	"github.com/fernando/financial-management-tracking/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestReportAPI_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping integration test: database not reachable at %s (%v)", dsn, err)
		return
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	tokenSvc := auth.NewJWTService("integration-test-secret-key-12345678", 1*time.Hour)
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, tokenSvc)

	accRepo := account.NewRepository(pool)
	accSvc := account.NewService(accRepo)

	catRepo := category.NewRepository(pool)
	catSvc := category.NewService(catRepo)

	txRepo := transaction.NewRepository(pool)
	txSvc := transaction.NewService(txRepo)

	budgetRepo := budget.NewRepository(pool)
	budgetSvc := budget.NewService(budgetRepo)

	recRepo := receivable.NewRepository(pool)
	recSvc := receivable.NewService(recRepo)

	invRepo := investment.NewRepository(pool)
	invSvc := investment.NewService(invRepo)

	repRepo := report.NewRepository(pool)
	repSvc := report.NewService(repRepo)
	repHandler := report.NewHandler(repSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/reports", repHandler.Routes())

	// Register test user
	uniqueEmail := fmt.Sprintf("repuser-%d@example.com", time.Now().UnixNano())
	regRes, err := authSvc.Register(ctx, uniqueEmail, "Password123!", 1)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	// Create test account
	acc, err := accSvc.CreateAccount(ctx, userID, account.CreateAccountInput{
		Name:           "Main Checking",
		Type:           account.TypeBank,
		OpeningBalance: 10000000,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Create test category
	cat, err := catSvc.CreateCategory(ctx, userID, category.CreateCategoryInput{
		Name: "Groceries",
		Type: "EXPENSE",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// Create transactions in September 2026
	// 1. Income
	_, err = txSvc.CreateTransaction(ctx, userID, transaction.CreateTransactionRequest{
		AccountID:       acc.ID,
		Type:            transaction.TypeIncome,
		Amount:          5000000,
		TransactionDate: "2026-09-05",
		Description:     "Bonus",
	})
	if err != nil {
		t.Fatalf("failed to create income: %v", err)
	}

	// 2. Expense
	_, err = txSvc.CreateTransaction(ctx, userID, transaction.CreateTransactionRequest{
		AccountID:       acc.ID,
		CategoryID:      &cat.ID,
		Type:            transaction.TypeExpense,
		Amount:          1500000,
		TransactionDate: "2026-09-12",
		Description:     "Supermarket",
	})
	if err != nil {
		t.Fatalf("failed to create expense: %v", err)
	}

	// 3. Create Budget for September 2026
	_, err = budgetSvc.CreateBudget(ctx, userID, budget.CreateBudgetInput{
		CategoryID:    cat.ID,
		CycleStart:    "2026-09-01",
		CycleEnd:      "2026-09-30",
		PlannedAmount: 2000000,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	// 4. Create Receivable
	dueDate := "2026-10-01"
	_, err = recSvc.Create(ctx, userID, receivable.CreateReceivableRequest{
		Counterparty:    "Alex",
		Principal:       1000000,
		DueDate:         &dueDate,
		SourceAccountID: &acc.ID,
	})
	if err != nil {
		t.Fatalf("failed to create receivable: %v", err)
	}

	// 5. Create Investment
	_, err = invSvc.Create(ctx, userID, investment.CreateInvestmentRequest{
		Name:            "S&P Index",
		Type:            investment.TypeMutualFund,
		Capital:         3000000,
		SourceAccountID: &acc.ID,
	})
	if err != nil {
		t.Fatalf("failed to create investment: %v", err)
	}

	// Test 1: Monthly Report
	reqMonthly, _ := http.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2026&month=09", nil)
	reqMonthly.Header.Set("Authorization", "Bearer "+token)
	rrMonthly := httptest.NewRecorder()
	r.ServeHTTP(rrMonthly, reqMonthly)

	if rrMonthly.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for monthly report, got %d: %s", rrMonthly.Code, rrMonthly.Body.String())
	}

	var monthlyEnv APIEnvelope
	if err := json.Unmarshal(rrMonthly.Body.Bytes(), &monthlyEnv); err != nil {
		t.Fatalf("failed to unmarshal monthly envelope: %v", err)
	}
	var monthlyReport report.ReportData
	if err := json.Unmarshal(monthlyEnv.Data, &monthlyReport); err != nil {
		t.Fatalf("failed to unmarshal monthly report data: %v", err)
	}

	if monthlyReport.Period != "2026-09" {
		t.Errorf("expected period 2026-09, got %s", monthlyReport.Period)
	}
	if monthlyReport.Income != 5000000 {
		t.Errorf("expected income 5000000, got %d", monthlyReport.Income)
	}
	if monthlyReport.Expense != 1500000 {
		t.Errorf("expected expense 1500000, got %d", monthlyReport.Expense)
	}
	if monthlyReport.NetCashFlow != 3500000 {
		t.Errorf("expected net cash flow 3500000, got %d", monthlyReport.NetCashFlow)
	}
	if len(monthlyReport.BudgetVsActual) != 1 {
		t.Fatalf("expected 1 budget item, got %d", len(monthlyReport.BudgetVsActual))
	}
	bItem := monthlyReport.BudgetVsActual[0]
	if bItem.Budget != 2000000 || bItem.Actual != 1500000 || bItem.Variance != 500000 {
		t.Errorf("unexpected budget item: %+v", bItem)
	}
	if monthlyReport.Assets.Receivables != 1000000 {
		t.Errorf("expected receivables 1000000, got %d", monthlyReport.Assets.Receivables)
	}
	if monthlyReport.Assets.Investments != 3000000 {
		t.Errorf("expected investments 3000000, got %d", monthlyReport.Assets.Investments)
	}
	if monthlyReport.NetWorth != monthlyReport.Assets.Total {
		t.Errorf("expected net worth to equal total assets, got net_worth=%d, total=%d", monthlyReport.NetWorth, monthlyReport.Assets.Total)
	}

	// Test 2: Cycle Report (date in September 2026)
	reqCycle, _ := http.NewRequest(http.MethodGet, "/api/v1/reports/cycle?date=2026-09-15", nil)
	reqCycle.Header.Set("Authorization", "Bearer "+token)
	rrCycle := httptest.NewRecorder()
	r.ServeHTTP(rrCycle, reqCycle)

	if rrCycle.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for cycle report, got %d: %s", rrCycle.Code, rrCycle.Body.String())
	}

	var cycleEnv APIEnvelope
	if err := json.Unmarshal(rrCycle.Body.Bytes(), &cycleEnv); err != nil {
		t.Fatalf("failed to unmarshal cycle envelope: %v", err)
	}
	var cycleReport report.ReportData
	if err := json.Unmarshal(cycleEnv.Data, &cycleReport); err != nil {
		t.Fatalf("failed to unmarshal cycle report data: %v", err)
	}

	// User default cycle_start_day is 1, so cycle for 2026-09-15 is 2026-09-01 to 2026-09-30
	if cycleReport.StartDate != "2026-09-01" || cycleReport.EndDate != "2026-09-30" {
		t.Errorf("unexpected cycle dates: %s to %s", cycleReport.StartDate, cycleReport.EndDate)
	}
	if cycleReport.Income != 5000000 || cycleReport.Expense != 1500000 {
		t.Errorf("unexpected cycle cash flows: income=%d, expense=%d", cycleReport.Income, cycleReport.Expense)
	}
}
