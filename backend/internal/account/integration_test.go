package account_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestAccountsAPI_Integration(t *testing.T) {
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

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/accounts", accountHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Register user
	userEmail := fmt.Sprintf("accuser-%s@example.com", uuid.New().String())
	regRes, err := authSvc.Register(ctx, userEmail, "Pass1234!", 1)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token

	// 1. Create Account
	createBody, _ := json.Marshal(map[string]interface{}{
		"name":            "Mandiri Savings",
		"type":            "BANK",
		"opening_balance": 5000000,
	})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/accounts", bytes.NewBuffer(createBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var env APIEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	var createdAcc account.Account
	_ = json.Unmarshal(env.Data, &createdAcc)

	if createdAcc.OpeningBalance != 5000000 || createdAcc.CurrentBalance != 5000000 {
		t.Errorf("unexpected balances in created account: %+v", createdAcc)
	}

	// 2. List Accounts
	listReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/accounts", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil {
		t.Fatalf("failed to list accounts: %v", err)
	}
	defer listResp.Body.Close()

	var listEnv APIEnvelope
	_ = json.NewDecoder(listResp.Body).Decode(&listEnv)
	var accounts []account.Account
	_ = json.Unmarshal(listEnv.Data, &accounts)

	if len(accounts) != 1 || accounts[0].ID != createdAcc.ID {
		t.Fatalf("expected 1 account with ID %v, got %d", createdAcc.ID, len(accounts))
	}

	// 3. Update Account
	updateBody, _ := json.Marshal(map[string]interface{}{
		"name":   "Mandiri Primary",
		"type":   "BANK",
		"status": "ACTIVE",
	})
	upReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/accounts/%s", ts.URL, createdAcc.ID), bytes.NewBuffer(updateBody))
	upReq.Header.Set("Authorization", "Bearer "+token)
	upReq.Header.Set("Content-Type", "application/json")
	upResp, err := http.DefaultClient.Do(upReq)
	if err != nil {
		t.Fatalf("failed to update account: %v", err)
	}
	defer upResp.Body.Close()
	if upResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on update, got %d", upResp.StatusCode)
	}

	// 4. Soft Delete Account
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/accounts/%s", ts.URL, createdAcc.ID), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("failed to delete account: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", delResp.StatusCode)
	}

	// 5. List after delete should be empty
	listReq2, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/accounts", nil)
	listReq2.Header.Set("Authorization", "Bearer "+token)
	listResp2, err := http.DefaultClient.Do(listReq2)
	if err != nil {
		t.Fatalf("failed to list accounts after delete: %v", err)
	}
	defer listResp2.Body.Close()
	var listEnv2 APIEnvelope
	_ = json.NewDecoder(listResp2.Body).Decode(&listEnv2)
	var accounts2 []account.Account
	_ = json.Unmarshal(listEnv2.Data, &accounts2)

	if len(accounts2) != 0 {
		t.Fatalf("expected 0 accounts after soft delete, got %d", len(accounts2))
	}
}
