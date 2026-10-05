package investment_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/investment"
	"github.com/go-chi/chi/v5"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestInvestmentAPI_Integration(t *testing.T) {
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

	invRepo := investment.NewRepository(pool)
	invSvc := investment.NewService(invRepo)
	invHandler := investment.NewHandler(invSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/investments", invHandler.Routes())

	// 1. Create a test user
	email := fmt.Sprintf("inv_user_%d@example.com", time.Now().UnixNano())
	regRes, err := authSvc.Register(ctx, email, "Password123!", 1)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	// 2. Create funding account with Rp 50,000,000
	fundAcc, err := accSvc.CreateAccount(ctx, userID, account.CreateAccountInput{
		Name:           "Stock Investment Fund",
		Type:           account.TypeInvestment,
		OpeningBalance: 50000000,
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	// 3. Create investment in BBCA with Rp 20,000,000 capital funded from account
	createPayload := investment.CreateInvestmentRequest{
		Name:            "BBCA - Bank Central Asia",
		Type:            investment.TypeStock,
		Capital:         20000000,
		SourceAccountID: &fundAcc.ID,
		Notes:           "Long term blue chip",
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/v1/investments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var createResp APIEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	var createdInv investment.Investment
	_ = json.Unmarshal(createResp.Data, &createdInv)

	if createdInv.Capital != 20000000 || createdInv.CurrentValue != 20000000 || createdInv.UnrealizedGain != 0 {
		t.Fatalf("unexpected investment fields: %+v", createdInv)
	}

	// Verify account balance deducted to Rp 30,000,000
	updatedAcc, _ := accSvc.GetAccount(ctx, userID, fundAcc.ID)
	if updatedAcc.CurrentBalance != 30000000 {
		t.Errorf("expected account balance 30000000, got %d", updatedAcc.CurrentBalance)
	}

	// 4. Update valuation to Rp 23,000,000 (+15%)
	valPayload := investment.UpdateValuationRequest{
		CurrentValue: 23000000,
	}
	valBody, _ := json.Marshal(valPayload)
	valReq := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/investments/%s", createdInv.ID), bytes.NewReader(valBody))
	valReq.Header.Set("Authorization", "Bearer "+token)
	valReq.Header.Set("Content-Type", "application/json")
	valW := httptest.NewRecorder()
	r.ServeHTTP(valW, valReq)

	if valW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for valuation update, got %d: %s", valW.Code, valW.Body.String())
	}

	// 5. Query portfolio summary
	listReq := httptest.NewRequest("GET", "/api/v1/investments", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for list, got %d", listW.Code)
	}

	var listResp APIEnvelope
	_ = json.Unmarshal(listW.Body.Bytes(), &listResp)
	var summary investment.PortfolioSummary
	_ = json.Unmarshal(listResp.Data, &summary)

	if summary.InvestmentsCount != 1 || summary.TotalCapital != 20000000 || summary.TotalCurrentValue != 23000000 || summary.TotalGainLoss != 3000000 || summary.TotalGainLossPercentage != 15.0 {
		t.Errorf("unexpected portfolio summary: %+v", summary)
	}
}
