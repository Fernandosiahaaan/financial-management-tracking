package allocation_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/allocation"
	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/fernando/financial-management-tracking/backend/internal/category"
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

func TestAllocationAPI_Integration(t *testing.T) {
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

	allocRepo := allocation.NewRepository(pool)
	allocSvc := allocation.NewService(allocRepo)
	allocHandler := allocation.NewHandler(allocSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/allocations", allocHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Register user
	userEmail := fmt.Sprintf("allocuser-%s@example.com", uuid.New().String())
	regRes, err := authSvc.Register(ctx, userEmail, "Pass1234!", 1)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	// 2. Create categories
	cat1, err := catSvc.CreateCategory(ctx, userID, category.CreateCategoryInput{
		Name:  "Living Expenses",
		Type:  category.TypeExpense,
		Icon:  "home",
		Color: "#3B82F6",
	})
	if err != nil {
		t.Fatalf("failed to create category 1: %v", err)
	}

	cat2, err := catSvc.CreateCategory(ctx, userID, category.CreateCategoryInput{
		Name:  "Savings Goal",
		Type:  category.TypeExpense,
		Icon:  "piggy-bank",
		Color: "#10B981",
	})
	if err != nil {
		t.Fatalf("failed to create category 2: %v", err)
	}

	// 3. Create account and record 10,000,000 income in cycle 2026-10-01 to 2026-10-31
	acc, err := accSvc.CreateAccount(ctx, userID, account.CreateAccountInput{
		Name:           "Payroll Account",
		Type:           account.TypeBank,
		OpeningBalance: 0,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	_, err = txSvc.CreateTransaction(ctx, userID, transaction.CreateTransactionRequest{
		AccountID:       acc.ID,
		Type:            transaction.TypeIncome,
		Amount:          10000000,
		TransactionDate: "2026-10-05",
		Description:     "October Salary",
	})
	if err != nil {
		t.Fatalf("failed to record income: %v", err)
	}

	// 4. Create Allocation 1: 4,000,000 (valid, within 10M income)
	createPayload1, _ := json.Marshal(allocation.CreateAllocationInput{
		CategoryID:      cat1.ID,
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 4000000,
	})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/allocations", bytes.NewBuffer(createPayload1))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create allocation: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var env APIEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	var createdAlloc allocation.Allocation
	_ = json.Unmarshal(env.Data, &createdAlloc)

	if createdAlloc.AllocatedAmount != 4000000 {
		t.Errorf("expected allocated amount 4000000, got %d", createdAlloc.AllocatedAmount)
	}

	// 5. Create Allocation 2 exceeding available income: 7,000,000 (4M + 7M = 11M > 10M)
	createPayload2, _ := json.Marshal(allocation.CreateAllocationInput{
		CategoryID:      cat2.ID,
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 7000000,
	})
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/allocations", bytes.NewBuffer(createPayload2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 Unprocessable Entity when exceeding income, got %d", resp2.StatusCode)
	}

	// 6. List Allocations for cycle
	listReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/allocations?cycle_start=2026-10-01&cycle_end=2026-10-31", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil {
		t.Fatalf("failed to list allocations: %v", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", listResp.StatusCode)
	}

	var listEnv APIEnvelope
	_ = json.NewDecoder(listResp.Body).Decode(&listEnv)
	var summary allocation.AllocationSummary
	_ = json.Unmarshal(listEnv.Data, &summary)

	if summary.TotalIncome != 10000000 {
		t.Errorf("expected total income 10000000, got %d", summary.TotalIncome)
	}
	if summary.TotalAllocated != 4000000 {
		t.Errorf("expected total allocated 4000000, got %d", summary.TotalAllocated)
	}
	if summary.RemainingIncome != 6000000 {
		t.Errorf("expected remaining income 6000000, got %d", summary.RemainingIncome)
	}
	if len(summary.Allocations) != 1 {
		t.Errorf("expected 1 allocation, got %d", len(summary.Allocations))
	}

	// 7. Update Allocation from 4,000,000 to 6,000,000
	updatePayload, _ := json.Marshal(allocation.UpdateAllocationInput{AllocatedAmount: 6000000})
	updateReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/allocations/%s", ts.URL, createdAlloc.ID), bytes.NewBuffer(updatePayload))
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		t.Fatalf("failed to update allocation: %v", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", updateResp.StatusCode)
	}

	// 8. Delete Allocation
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/allocations/%s", ts.URL, createdAlloc.ID), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("failed to delete allocation: %v", err)
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", delResp.StatusCode)
	}
}
