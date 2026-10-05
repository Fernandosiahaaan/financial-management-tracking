package dashboard_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/dashboard"
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/fernando/financial-management-tracking/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestDashboardAPI_Integration(t *testing.T) {
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

	catRepo := category.NewRepository(pool)
	catSvc := category.NewService(catRepo)

	accRepo := account.NewRepository(pool)
	accSvc := account.NewService(accRepo)

	txRepo := transaction.NewRepository(pool)
	txSvc := transaction.NewService(txRepo)

	budgetRepo := budget.NewRepository(pool)
	budgetSvc := budget.NewService(budgetRepo)

	dashRepo := dashboard.NewRepository(pool)
	dashSvc := dashboard.NewService(dashRepo)
	dashHandler := dashboard.NewHandler(dashSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/dashboard", dashHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Register user with cycle_start_day = 1
	userEmail := fmt.Sprintf("dashuser-%s@example.com", uuid.New().String())
	regRes, err := authSvc.Register(ctx, userEmail, "Pass1234!", 1)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	// 2. Create Account with opening balance of 10,000,000
	acc, err := accSvc.CreateAccount(ctx, userID, account.CreateAccountInput{
		Name:           "Mandiri Savings",
		Type:           account.TypeBank,
		OpeningBalance: 10000000,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// 3. Create Categories
	catGroceries, err := catSvc.CreateCategory(ctx, userID, category.CreateCategoryInput{
		Name:  "Groceries",
		Type:  category.TypeExpense,
		Icon:  "basket",
		Color: "#EF4444",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// 4. Record transactions in previous cycle (September 2026: 2026-09-01 to 2026-09-30)
	// September Income: +5,000,000
	_, err = txSvc.CreateTransaction(ctx, userID, transaction.CreateTransactionRequest{
		AccountID:       acc.ID,
		Type:            transaction.TypeIncome,
		Amount:          5000000,
		TransactionDate: "2026-09-15",
		Description:     "September Bonus",
	})
	if err != nil {
		t.Fatalf("failed to record previous cycle income: %v", err)
	}

	// At end of previous cycle (2026-09-30):
	// Total Assets = Opening (10M) + Sep Income (5M) = 15M

	// 5. Record transactions in current cycle (October 2026: 2026-10-01 to 2026-10-31)
	// October Income: +12,000,000
	_, err = txSvc.CreateTransaction(ctx, userID, transaction.CreateTransactionRequest{
		AccountID:       acc.ID,
		Type:            transaction.TypeIncome,
		Amount:          12000000,
		TransactionDate: "2026-10-05",
		Description:     "October Salary",
	})
	if err != nil {
		t.Fatalf("failed to record current cycle income: %v", err)
	}

	// October Expense: -3,500,000 in Groceries
	_, err = txSvc.CreateTransaction(ctx, userID, transaction.CreateTransactionRequest{
		AccountID:       acc.ID,
		CategoryID:      &catGroceries.ID,
		Type:            transaction.TypeExpense,
		Amount:          3500000,
		TransactionDate: "2026-10-10",
		Description:     "Monthly Groceries",
	})
	if err != nil {
		t.Fatalf("failed to record current cycle expense: %v", err)
	}

	// 6. Create Budget in October for Groceries: Planned 3,000,000
	// Actual spending will be 3,500,000 -> Variance -500,000 (Overspent)
	_, err = budgetSvc.CreateBudget(ctx, userID, budget.CreateBudgetInput{
		CategoryID:    catGroceries.ID,
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 3000000,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	// 7. Test GET /api/v1/dashboard?date=2026-10-15
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/dashboard?date=2026-10-15", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var env APIEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	var data dashboard.DashboardData
	_ = json.Unmarshal(env.Data, &data)

	// Check cycle
	if data.Cycle.StartDate != "2026-10-01" || data.Cycle.EndDate != "2026-10-31" {
		t.Errorf("unexpected cycle dates: %s to %s", data.Cycle.StartDate, data.Cycle.EndDate)
	}

	// Check cash flow
	if data.Income != 12000000 {
		t.Errorf("expected income 12000000, got %d", data.Income)
	}
	if data.Expense != 3500000 {
		t.Errorf("expected expense 3500000, got %d", data.Expense)
	}
	if data.NetCashFlow != 8500000 {
		t.Errorf("expected net cash flow 8500000, got %d", data.NetCashFlow)
	}

	// Check assets
	// Total Assets = 10M opening + 5M Sep + 12M Oct - 3.5M Oct = 23.5M
	expectedTotalAssets := int64(23500000)
	if data.TotalAssets != expectedTotalAssets {
		t.Errorf("expected total assets %d, got %d", expectedTotalAssets, data.TotalAssets)
	}
	// Previous Cycle Assets (as of 2026-09-30) = 15M
	expectedPrevAssets := int64(15000000)
	if data.PreviousCycleAssets != expectedPrevAssets {
		t.Errorf("expected previous cycle assets %d, got %d", expectedPrevAssets, data.PreviousCycleAssets)
	}
	// Asset growth = 23.5M - 15M = 8.5M (matching October net cashflow)
	expectedGrowth := int64(8500000)
	if data.AssetGrowth != expectedGrowth {
		t.Errorf("expected asset growth %d, got %d", expectedGrowth, data.AssetGrowth)
	}

	// Check budget progress
	if len(data.Budgets) != 1 {
		t.Fatalf("expected 1 budget, got %d", len(data.Budgets))
	}
	b := data.Budgets[0]
	if b.Planned != 3000000 || b.Actual != 3500000 || b.Variance != -500000 || !b.Overspent {
		t.Errorf("unexpected budget details: %+v", b)
	}

	// Check accounts
	if len(data.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(data.Accounts))
	}
	if data.Accounts[0].Balance != expectedTotalAssets {
		t.Errorf("expected account balance %d, got %d", expectedTotalAssets, data.Accounts[0].Balance)
	}
}
