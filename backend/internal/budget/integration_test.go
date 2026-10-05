package budget_test

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

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/fernando/financial-management-tracking/backend/internal/budget"
	"github.com/fernando/financial-management-tracking/backend/internal/category"
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestBudgetAPI_Integration(t *testing.T) {
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

	budgetRepo := budget.NewRepository(pool)
	budgetSvc := budget.NewService(budgetRepo)
	budgetHandler := budget.NewHandler(budgetSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/budgets", budgetHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Register user 1
	user1Email := fmt.Sprintf("budgetuser1-%s@example.com", uuid.New().String())
	regRes1, err := authSvc.Register(ctx, user1Email, "Pass1234!", 1)
	if err != nil {
		t.Fatalf("failed to register user 1: %v", err)
	}
	token1 := regRes1.Token
	user1ID := regRes1.Profile.ID

	// 2. Create category for user 1
	cat1, err := catSvc.CreateCategory(ctx, user1ID, category.CreateCategoryInput{
		Name:  "Groceries",
		Type:  category.TypeExpense,
		Icon:  "basket",
		Color: "#EF4444",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	// 3. Create Budget
	createPayload, _ := json.Marshal(budget.CreateBudgetInput{
		CategoryID:    cat1.ID,
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 3000000,
	})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/budgets", bytes.NewBuffer(createPayload))
	req.Header.Set("Authorization", "Bearer "+token1)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}
	defer resp.Body.Close()

	var env APIEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d (msg: %s)", resp.StatusCode, env.Message)
	}
	var createdBudget budget.Budget
	_ = json.Unmarshal(env.Data, &createdBudget)

	if createdBudget.PlannedAmount != 3000000 {
		t.Errorf("expected planned amount 3000000, got %d", createdBudget.PlannedAmount)
	}
	if createdBudget.CategoryID != cat1.ID {
		t.Errorf("expected category_id %s, got %s", cat1.ID, createdBudget.CategoryID)
	}

	// 4. Duplicate budget for same category/cycle should return 409 Conflict
	dupReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/budgets", bytes.NewBuffer(createPayload))
	dupReq.Header.Set("Authorization", "Bearer "+token1)
	dupReq.Header.Set("Content-Type", "application/json")
	dupResp, err := http.DefaultClient.Do(dupReq)
	if err != nil {
		t.Fatalf("failed to send dup request: %v", err)
	}
	defer dupResp.Body.Close()

	if dupResp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 Conflict for duplicate budget, got %d", dupResp.StatusCode)
	}

	// 5. List Budgets for cycle
	listReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/budgets?cycle_start=2026-10-01&cycle_end=2026-10-31", nil)
	listReq.Header.Set("Authorization", "Bearer "+token1)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil {
		t.Fatalf("failed to list budgets: %v", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", listResp.StatusCode)
	}

	var listEnv APIEnvelope
	_ = json.NewDecoder(listResp.Body).Decode(&listEnv)
	var budgets []budget.Budget
	_ = json.Unmarshal(listEnv.Data, &budgets)

	if len(budgets) != 1 {
		t.Fatalf("expected 1 budget in list, got %d", len(budgets))
	}
	if budgets[0].CategoryName != "Groceries" {
		t.Errorf("expected joined category name 'Groceries', got '%s'", budgets[0].CategoryName)
	}
	// Variance = planned (3M) - actual (0) = 3M
	if budgets[0].Variance != 3000000 {
		t.Errorf("expected variance 3000000, got %d", budgets[0].Variance)
	}

	// 6. Update Budget
	updatePayload, _ := json.Marshal(budget.UpdateBudgetInput{PlannedAmount: 4500000})
	updateReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/budgets/%s", ts.URL, createdBudget.ID), bytes.NewBuffer(updatePayload))
	updateReq.Header.Set("Authorization", "Bearer "+token1)
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		t.Fatalf("failed to update budget: %v", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", updateResp.StatusCode)
	}

	// 7. Delete Budget (Soft-delete)
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/budgets/%s", ts.URL, createdBudget.ID), nil)
	delReq.Header.Set("Authorization", "Bearer "+token1)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("failed to delete budget: %v", err)
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", delResp.StatusCode)
	}

	// 8. Re-list budgets should now be empty
	listReq2, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/budgets?cycle_start=2026-10-01&cycle_end=2026-10-31", nil)
	listReq2.Header.Set("Authorization", "Bearer "+token1)
	listResp2, err := http.DefaultClient.Do(listReq2)
	if err != nil {
		t.Fatalf("failed to list budgets after delete: %v", err)
	}
	defer listResp2.Body.Close()

	var listEnv2 APIEnvelope
	_ = json.NewDecoder(listResp2.Body).Decode(&listEnv2)
	var budgets2 []budget.Budget
	_ = json.Unmarshal(listEnv2.Data, &budgets2)

	if len(budgets2) != 0 {
		t.Errorf("expected 0 budgets after soft delete, got %d", len(budgets2))
	}
}
