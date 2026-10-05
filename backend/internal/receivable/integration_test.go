package receivable_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/receivable"
	"github.com/go-chi/chi/v5"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestReceivableAPI_Integration(t *testing.T) {
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

	recRepo := receivable.NewRepository(pool)
	recSvc := receivable.NewService(recRepo)
	recHandler := receivable.NewHandler(recSvc)

	r := chi.NewRouter()
	r.Use(auth.RequireAuth(tokenSvc))
	r.Mount("/api/v1/receivables", recHandler.Routes())

	// 1. Create a unique test user
	email := fmt.Sprintf("rec_user_%d@example.com", time.Now().UnixNano())
	regRes, err := authSvc.Register(ctx, email, "Password123!", 1)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	token := regRes.Token
	userID := regRes.Profile.ID

	// 2. Create source account with Rp 10,000,000 balance
	sourceAcc, err := accSvc.CreateAccount(ctx, userID, account.CreateAccountInput{
		Name:           "BCA Checking",
		Type:           account.TypeBank,
		OpeningBalance: 10000000,
	})
	if err != nil {
		t.Fatalf("failed to create source account: %v", err)
	}

	// 3. Create target repayment account with Rp 2,000,000 balance
	targetAcc, err := accSvc.CreateAccount(ctx, userID, account.CreateAccountInput{
		Name:           "Cash Wallet",
		Type:           account.TypeCash,
		OpeningBalance: 2000000,
	})
	if err != nil {
		t.Fatalf("failed to create target account: %v", err)
	}

	// 4. Create a receivable lending Rp 4,000,000 to Andi
	createPayload := receivable.CreateReceivableRequest{
		Counterparty:    "Andi",
		Principal:       4000000,
		SourceAccountID: &sourceAcc.ID,
		Notes:           "Friendly loan",
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/v1/receivables", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var recResp APIEnvelope
	_ = json.Unmarshal(w.Body.Bytes(), &recResp)
	var createdRec receivable.Receivable
	_ = json.Unmarshal(recResp.Data, &createdRec)

	if createdRec.Principal != 4000000 || createdRec.RemainingAmount != 4000000 || createdRec.Status != receivable.StatusActive {
		t.Fatalf("unexpected created receivable: %+v", createdRec)
	}

	// Verify source account balance decreased to Rp 6,000,000
	updatedSourceAcc, _ := accSvc.GetAccount(ctx, userID, sourceAcc.ID)
	if updatedSourceAcc.CurrentBalance != 6000000 {
		t.Errorf("expected source account balance 6000000, got %d", updatedSourceAcc.CurrentBalance)
	}

	// 5. Record partial repayment of Rp 1,500,000 into Cash Wallet
	payPayload := receivable.RecordPaymentRequest{
		Amount:          1500000,
		TargetAccountID: targetAcc.ID,
		PaymentDate:     "2026-10-15",
		Notes:           "First installment",
	}
	payBody, _ := json.Marshal(payPayload)
	payReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/receivables/%s/payments", createdRec.ID), bytes.NewReader(payBody))
	payReq.Header.Set("Authorization", "Bearer "+token)
	payReq.Header.Set("Content-Type", "application/json")
	payW := httptest.NewRecorder()
	r.ServeHTTP(payW, payReq)

	if payW.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for payment, got %d: %s", payW.Code, payW.Body.String())
	}

	// Verify target account balance increased to Rp 3,500,000
	updatedTargetAcc, _ := accSvc.GetAccount(ctx, userID, targetAcc.ID)
	if updatedTargetAcc.CurrentBalance != 3500000 {
		t.Errorf("expected target account balance 3500000, got %d", updatedTargetAcc.CurrentBalance)
	}

	// 6. Test overpayment rejection: remaining is 2,500,000, attempt to pay 3,000,000
	overpayPayload := receivable.RecordPaymentRequest{
		Amount:          3000000,
		TargetAccountID: targetAcc.ID,
		PaymentDate:     "2026-10-16",
	}
	overBody, _ := json.Marshal(overpayPayload)
	overReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/receivables/%s/payments", createdRec.ID), bytes.NewReader(overBody))
	overReq.Header.Set("Authorization", "Bearer "+token)
	overReq.Header.Set("Content-Type", "application/json")
	overW := httptest.NewRecorder()
	r.ServeHTTP(overW, overReq)

	if overW.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 Unprocessable Entity for overpayment, got %d", overW.Code)
	}

	// 7. Pay off remaining Rp 2,500,000
	finalPayPayload := receivable.RecordPaymentRequest{
		Amount:          2500000,
		TargetAccountID: targetAcc.ID,
		PaymentDate:     "2026-10-16",
	}
	finalBody, _ := json.Marshal(finalPayPayload)
	finalReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/receivables/%s/payments", createdRec.ID), bytes.NewReader(finalBody))
	finalReq.Header.Set("Authorization", "Bearer "+token)
	finalReq.Header.Set("Content-Type", "application/json")
	finalW := httptest.NewRecorder()
	r.ServeHTTP(finalW, finalReq)

	if finalW.Code != http.StatusCreated {
		t.Fatalf("expected 201 for final payment, got %d", finalW.Code)
	}

	// Verify status is now PAID
	getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/receivables/%s", createdRec.ID), nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	var getResp APIEnvelope
	_ = json.Unmarshal(getW.Body.Bytes(), &getResp)
	var finalRec receivable.Receivable
	_ = json.Unmarshal(getResp.Data, &finalRec)

	if finalRec.Status != receivable.StatusPaid || finalRec.RemainingAmount != 0 || len(finalRec.Payments) != 2 {
		t.Errorf("expected PAID status with 0 remaining and 2 payments, got: %+v", finalRec)
	}
}
