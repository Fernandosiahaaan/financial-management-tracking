package transaction_test

import (
	"bytes"
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
	"github.com/fernando/financial-management-tracking/backend/internal/category"
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/fernando/financial-management-tracking/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestTransactionsAPI_Integration(t *testing.T) {
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
	accountHandler := account.NewHandler(accountSvc)

	categoryRepo := category.NewRepository(pool)
	categorySvc := category.NewService(categoryRepo)
	categoryHandler := category.NewHandler(categorySvc)

	txRepo := transaction.NewRepository(pool)
	txSvc := transaction.NewService(txRepo)
	txHandler := transaction.NewHandler(txSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/accounts", accountHandler.Routes())
	r.Mount("/api/v1/categories", categoryHandler.Routes())
	r.Mount("/api/v1/transactions", txHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Create a test user
	uniqueEmail := fmt.Sprintf("txtest_%d@example.com", time.Now().UnixNano())
	regRes, err := authSvc.Register(context.Background(), uniqueEmail, "SecurePassword123!", 1)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	// Helper function for authorized requests
	doAuthReq := func(method, path string, body any) (*http.Response, APIEnvelope) {
		var reqBody []byte
		if body != nil {
			reqBody, _ = json.Marshal(body)
		}
		req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("failed to build request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to perform request %s %s: %v", method, path, err)
		}

		var env APIEnvelope
		if res.StatusCode != http.StatusNoContent && res.Body != nil {
			_ = json.NewDecoder(res.Body).Decode(&env)
		}
		return res, env
	}

	// 2. Create Accounts: Bank (1,000,000) and Cash (500,000)
	bankAcc, err := accountSvc.CreateAccount(context.Background(), userID, account.CreateAccountInput{
		Name:           "BCA Checking",
		Type:           account.TypeBank,
		OpeningBalance: 1000000,
	})
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	cashAcc, err := accountSvc.CreateAccount(context.Background(), userID, account.CreateAccountInput{
		Name:           "Cash Wallet",
		Type:           account.TypeCash,
		OpeningBalance: 500000,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}

	// 3. Create Category: Food (Expense)
	foodCat, err := categorySvc.CreateCategory(context.Background(), userID, category.CreateCategoryInput{
		Name: "Food & Dining",
		Type: category.TypeExpense,
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// 4. Record EXPENSE: 200,000 from Bank
	res, env := doAuthReq("POST", "/api/v1/transactions", transaction.CreateTransactionRequest{
		AccountID:       bankAcc.ID,
		CategoryID:      &foodCat.ID,
		Type:            transaction.TypeExpense,
		Amount:          200000,
		TransactionDate: "2026-10-01",
		Description:     "Dinner with friends",
	})
	if res.StatusCode != http.StatusCreated || !env.Success {
		t.Fatalf("failed to record expense: status=%d, msg=%s", res.StatusCode, env.Message)
	}
	var expenseTx transaction.Transaction
	_ = json.Unmarshal(env.Data, &expenseTx)
	if expenseTx.Amount != 200000 || expenseTx.AccountName != "BCA Checking" {
		t.Errorf("unexpected expense transaction data: %+v", expenseTx)
	}

	// Verify Bank balance: 1,000,000 - 200,000 = 800,000
	updatedBank, _ := accountSvc.GetAccount(context.Background(), userID, bankAcc.ID)
	if updatedBank.CurrentBalance != 800000 {
		t.Fatalf("expected bank balance 800000, got %d", updatedBank.CurrentBalance)
	}

	// 5. Record INCOME: 300,000 to Cash
	res, env = doAuthReq("POST", "/api/v1/transactions", transaction.CreateTransactionRequest{
		AccountID:       cashAcc.ID,
		Type:            transaction.TypeIncome,
		Amount:          300000,
		TransactionDate: "2026-10-02",
		Description:     "Freelance cash gig",
	})
	if res.StatusCode != http.StatusCreated || !env.Success {
		t.Fatalf("failed to record income: status=%d", res.StatusCode)
	}

	// Verify Cash balance: 500,000 + 300,000 = 800,000
	updatedCash, _ := accountSvc.GetAccount(context.Background(), userID, cashAcc.ID)
	if updatedCash.CurrentBalance != 800000 {
		t.Fatalf("expected cash balance 800000, got %d", updatedCash.CurrentBalance)
	}

	// 6. Record TRANSFER: 100,000 from Bank to Cash
	res, env = doAuthReq("POST", "/api/v1/transactions", transaction.CreateTransactionRequest{
		AccountID:            bankAcc.ID,
		DestinationAccountID: &cashAcc.ID,
		Type:                 transaction.TypeTransfer,
		Amount:               100000,
		TransactionDate:      "2026-10-03",
		Description:          "ATM withdrawal to wallet",
	})
	if res.StatusCode != http.StatusCreated || !env.Success {
		t.Fatalf("failed to record transfer: status=%d", res.StatusCode)
	}

	var transferTx transaction.Transaction
	_ = json.Unmarshal(env.Data, &transferTx)

	// Verify Bank: 800,000 - 100,000 = 700,000
	// Verify Cash: 800,000 + 100,000 = 900,000
	updatedBank, _ = accountSvc.GetAccount(context.Background(), userID, bankAcc.ID)
	if updatedBank.CurrentBalance != 700000 {
		t.Fatalf("expected bank balance 700000, got %d", updatedBank.CurrentBalance)
	}
	updatedCash, _ = accountSvc.GetAccount(context.Background(), userID, cashAcc.ID)
	if updatedCash.CurrentBalance != 900000 {
		t.Fatalf("expected cash balance 900000, got %d", updatedCash.CurrentBalance)
	}

	// 7. Verify List and Filter
	res, env = doAuthReq("GET", "/api/v1/transactions?type=TRANSFER", nil)
	if res.StatusCode != http.StatusOK || !env.Success {
		t.Fatalf("failed to list transfer transactions: status=%d", res.StatusCode)
	}
	var transfers []transaction.Transaction
	_ = json.Unmarshal(env.Data, &transfers)
	if len(transfers) != 1 || transfers[0].ID != transferTx.ID {
		t.Fatalf("expected 1 transfer transaction, got %d", len(transfers))
	}

	// 8. Update Transaction: change transfer from 100,000 to 150,000
	res, env = doAuthReq("PUT", fmt.Sprintf("/api/v1/transactions/%s", transferTx.ID), transaction.UpdateTransactionRequest{
		AccountID:            bankAcc.ID,
		DestinationAccountID: &cashAcc.ID,
		Type:                 transaction.TypeTransfer,
		Amount:               150000,
		TransactionDate:      "2026-10-03",
		Description:          "ATM withdrawal updated",
	})
	if res.StatusCode != http.StatusOK || !env.Success {
		t.Fatalf("failed to update transfer transaction: status=%d", res.StatusCode)
	}

	// Verify balances recalculated:
	// Bank: 800,000 - 150,000 = 650,000
	// Cash: 800,000 + 150,000 = 950,000
	updatedBank, _ = accountSvc.GetAccount(context.Background(), userID, bankAcc.ID)
	if updatedBank.CurrentBalance != 650000 {
		t.Fatalf("expected bank balance 650000 after update, got %d", updatedBank.CurrentBalance)
	}
	updatedCash, _ = accountSvc.GetAccount(context.Background(), userID, cashAcc.ID)
	if updatedCash.CurrentBalance != 950000 {
		t.Fatalf("expected cash balance 950000 after update, got %d", updatedCash.CurrentBalance)
	}

	// 9. Soft-delete Transfer Transaction -> balances should revert back to before the transfer:
	// Bank: 800,000, Cash: 800,000
	res, _ = doAuthReq("DELETE", fmt.Sprintf("/api/v1/transactions/%s", transferTx.ID), nil)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("failed to delete transfer transaction: status=%d", res.StatusCode)
	}

	updatedBank, _ = accountSvc.GetAccount(context.Background(), userID, bankAcc.ID)
	if updatedBank.CurrentBalance != 800000 {
		t.Fatalf("expected bank balance 800000 after delete, got %d", updatedBank.CurrentBalance)
	}
	updatedCash, _ = accountSvc.GetAccount(context.Background(), userID, cashAcc.ID)
	if updatedCash.CurrentBalance != 800000 {
		t.Fatalf("expected cash balance 800000 after delete, got %d", updatedCash.CurrentBalance)
	}

	// 10. Multi-user isolation test: another user cannot see or modify this transaction
	otherUser, _ := authSvc.Register(context.Background(), fmt.Sprintf("other_%d@example.com", time.Now().UnixNano()), "SecurePassword123!", 1)
	otherReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/transactions/%s", ts.URL, expenseTx.ID), nil)
	otherReq.Header.Set("Authorization", "Bearer "+otherUser.Token)
	otherRes, err := http.DefaultClient.Do(otherReq)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	if otherRes.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for other user viewing transaction, got %d", otherRes.StatusCode)
	}
}
