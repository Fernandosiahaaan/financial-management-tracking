package cycle_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/cycle"
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/fernando/financial-management-tracking/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestCycleAPI_Integration(t *testing.T) {
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

	accountRepo := account.NewRepository(pool)
	accountSvc := account.NewService(accountRepo)

	txRepo := transaction.NewRepository(pool)
	txSvc := transaction.NewService(txRepo)

	cycleRepo := cycle.NewRepository(pool)
	cycleSvc := cycle.NewService(cycleRepo)
	cycleHandler := cycle.NewHandler(cycleSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/cycle", cycleHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Register test user with cycle_start_day = 25
	uniqueEmail := fmt.Sprintf("cycletest_%d@example.com", time.Now().UnixNano())
	regRes, err := authSvc.Register(context.Background(), uniqueEmail, "SecurePassword123!", 25)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	doAuthGet := func(path string) (*http.Response, APIEnvelope) {
		req, err := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to perform request: %v", err)
		}

		var env APIEnvelope
		if res.Body != nil {
			_ = json.NewDecoder(res.Body).Decode(&env)
		}
		return res, env
	}

	// 2. Test /api/v1/cycle/current with target date 2026-08-30
	res, env := doAuthGet("/api/v1/cycle/current?date=2026-08-30")
	if res.StatusCode != http.StatusOK || !env.Success {
		t.Fatalf("failed to get current cycle: status=%d", res.StatusCode)
	}

	var cycleInfo cycle.CycleInfo
	_ = json.Unmarshal(env.Data, &cycleInfo)
	if cycleInfo.CycleStartDay != 25 {
		t.Errorf("expected cycle_start_day 25, got %d", cycleInfo.CycleStartDay)
	}
	if cycleInfo.StartDate != "2026-08-25" || cycleInfo.EndDate != "2026-09-24" {
		t.Errorf("expected 2026-08-25 to 2026-09-24, got [%s, %s]", cycleInfo.StartDate, cycleInfo.EndDate)
	}
	if cycleInfo.DayOfCycle != 6 { // August 25 is day 1, Aug 30 is day 6
		t.Errorf("expected dayOfCycle 6, got %d", cycleInfo.DayOfCycle)
	}

	// 3. Create two accounts and seed transactions
	acc1, err := accountSvc.CreateAccount(context.Background(), userID, account.CreateAccountInput{
		Name:           "Cycle Bank",
		Type:           account.TypeBank,
		OpeningBalance: 0,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	acc2, err := accountSvc.CreateAccount(context.Background(), userID, account.CreateAccountInput{
		Name:           "Cycle Cash",
		Type:           account.TypeCash,
		OpeningBalance: 0,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// Tx 1: Income 10,000,000 on 2026-08-26 (inside cycle)
	_, err = txSvc.CreateTransaction(context.Background(), userID, transaction.CreateTransactionRequest{
		AccountID:       acc1.ID,
		Type:            transaction.TypeIncome,
		Amount:          10000000,
		TransactionDate: "2026-08-26",
		Description:     "Salary",
	})
	if err != nil {
		t.Fatalf("failed to create income: %v", err)
	}

	// Tx 2: Expense 2,000,000 on 2026-09-02 (inside cycle)
	_, err = txSvc.CreateTransaction(context.Background(), userID, transaction.CreateTransactionRequest{
		AccountID:       acc1.ID,
		Type:            transaction.TypeExpense,
		Amount:          2000000,
		TransactionDate: "2026-09-02",
		Description:     "Groceries",
	})
	if err != nil {
		t.Fatalf("failed to create expense: %v", err)
	}

	// Tx 3: Transfer 1,000,000 from Bank to Cash on 2026-09-10 (inside cycle, should NOT affect income/expense)
	_, err = txSvc.CreateTransaction(context.Background(), userID, transaction.CreateTransactionRequest{
		AccountID:            acc1.ID,
		DestinationAccountID: &acc2.ID,
		Type:                 transaction.TypeTransfer,
		Amount:               1000000,
		TransactionDate:      "2026-09-10",
		Description:          "Wallet top up",
	})
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}

	// Tx 4: Expense 500,000 on 2026-09-28 (OUTSIDE cycle, belongs to next cycle: 2026-09-25 to 2026-10-24)
	_, err = txSvc.CreateTransaction(context.Background(), userID, transaction.CreateTransactionRequest{
		AccountID:       acc1.ID,
		Type:            transaction.TypeExpense,
		Amount:          500000,
		TransactionDate: "2026-09-28",
		Description:     "Next cycle expense",
	})
	if err != nil {
		t.Fatalf("failed to create next cycle expense: %v", err)
	}

	// 4. Test /api/v1/cycle/summary for 2026-08-30
	res, env = doAuthGet("/api/v1/cycle/summary?date=2026-08-30")
	if res.StatusCode != http.StatusOK || !env.Success {
		t.Fatalf("failed to get cycle summary: status=%d", res.StatusCode)
	}

	var summary cycle.CycleSummary
	_ = json.Unmarshal(env.Data, &summary)

	if summary.TotalIncome != 10000000 {
		t.Errorf("expected total income 10000000, got %d", summary.TotalIncome)
	}
	if summary.TotalExpense != 2000000 {
		t.Errorf("expected total expense 2000000, got %d", summary.TotalExpense)
	}
	if summary.NetSavings != 8000000 {
		t.Errorf("expected net savings 8000000, got %d", summary.NetSavings)
	}
	if summary.TransactionCount != 3 {
		t.Errorf("expected transaction count 3, got %d", summary.TransactionCount)
	}

	// 5. Test /api/v1/cycle/summary for 2026-09-28 (next cycle)
	res, env = doAuthGet("/api/v1/cycle/summary?date=2026-09-28")
	if res.StatusCode != http.StatusOK || !env.Success {
		t.Fatalf("failed to get next cycle summary: status=%d", res.StatusCode)
	}

	var nextSummary cycle.CycleSummary
	_ = json.Unmarshal(env.Data, &nextSummary)

	if nextSummary.Cycle.StartDate != "2026-09-25" || nextSummary.Cycle.EndDate != "2026-10-24" {
		t.Errorf("expected next cycle [2026-09-25, 2026-10-24], got [%s, %s]",
			nextSummary.Cycle.StartDate, nextSummary.Cycle.EndDate)
	}
	if nextSummary.TotalIncome != 0 {
		t.Errorf("expected next cycle income 0, got %d", nextSummary.TotalIncome)
	}
	if nextSummary.TotalExpense != 500000 {
		t.Errorf("expected next cycle expense 500000, got %d", nextSummary.TotalExpense)
	}
	if nextSummary.NetSavings != -500000 {
		t.Errorf("expected next cycle net savings -500000, got %d", nextSummary.NetSavings)
	}
	if nextSummary.TransactionCount != 1 {
		t.Errorf("expected next cycle count 1, got %d", nextSummary.TransactionCount)
	}
}
