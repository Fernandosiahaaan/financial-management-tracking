package auth_test

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
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type APIEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func TestAuthEndpoints_Integration(t *testing.T) {
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

	// Apply migrations
	if err := database.Migrate(ctx, pool, "../../migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	tokenSvc := auth.NewJWTService("integration-test-secret-key-12345678", 1*time.Hour)
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, tokenSvc)
	authHandler := auth.NewHandler(authSvc, tokenSvc)

	r := chi.NewRouter()
	r.Mount("/api/v1/auth", authHandler.Routes())

	ts := httptest.NewServer(r)
	defer ts.Close()

	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String())
	password := "SecurePassword123!"

	// 1. Test Register
	regBody, _ := json.Marshal(map[string]interface{}{
		"email":           uniqueEmail,
		"password":        password,
		"cycle_start_day": 25,
	})

	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(regBody))
	if err != nil {
		t.Fatalf("failed to send register request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created on register, got %d", resp.StatusCode)
	}

	var regEnvelope APIEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&regEnvelope); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	if !regEnvelope.Success {
		t.Fatalf("expected register success true")
	}

	var authData auth.AuthResponse
	if err := json.Unmarshal(regEnvelope.Data, &authData); err != nil {
		t.Fatalf("failed to unmarshal auth data: %v", err)
	}
	if authData.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if authData.Profile.CycleStartDay != 25 {
		t.Errorf("expected cycle start day 25, got %d", authData.Profile.CycleStartDay)
	}

	// 2. Test Duplicate Register -> 409 Conflict
	dupResp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(regBody))
	if err != nil {
		t.Fatalf("failed duplicate register request: %v", err)
	}
	defer dupResp.Body.Close()

	if dupResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on duplicate register, got %d", dupResp.StatusCode)
	}

	// 3. Test Login with valid credentials -> 200 OK
	loginBody, _ := json.Marshal(map[string]interface{}{
		"email":    uniqueEmail,
		"password": password,
	})

	loginResp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("failed login request: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK on login, got %d", loginResp.StatusCode)
	}

	var loginEnvelope APIEnvelope
	_ = json.NewDecoder(loginResp.Body).Decode(&loginEnvelope)
	var loginData auth.AuthResponse
	_ = json.Unmarshal(loginEnvelope.Data, &loginData)

	authToken := loginData.Token
	if authToken == "" {
		t.Fatal("expected token on login")
	}

	// 4. Test Login with wrong password -> 401 Unauthorized
	wrongPassBody, _ := json.Marshal(map[string]interface{}{
		"email":    uniqueEmail,
		"password": "WrongPassword!",
	})
	wrongResp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(wrongPassBody))
	if err != nil {
		t.Fatalf("failed to post login: %v", err)
	}
	defer wrongResp.Body.Close()
	if wrongResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong password, got %d", wrongResp.StatusCode)
	}

	// 5. Test Protected GET /api/v1/auth/me without token -> 401
	unauthReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/auth/me", nil)
	unauthResp, err := http.DefaultClient.Do(unauthReq)
	if err != nil {
		t.Fatalf("failed to call unauth me: %v", err)
	}
	defer unauthResp.Body.Close()
	if unauthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 on unauthenticated me request, got %d", unauthResp.StatusCode)
	}

	// 6. Test Protected GET /api/v1/auth/me with valid Bearer token -> 200 OK
	authReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/auth/me", nil)
	authReq.Header.Set("Authorization", "Bearer "+authToken)
	meResp, err := http.DefaultClient.Do(authReq)
	if err != nil {
		t.Fatalf("failed to call /auth/me: %v", err)
	}
	defer meResp.Body.Close()
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /auth/me, got %d", meResp.StatusCode)
	}

	// 7. Test PUT /api/v1/auth/settings -> 200 OK
	updateBody, _ := json.Marshal(map[string]interface{}{
		"cycle_start_day": 10,
	})
	updateReq, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/auth/settings", bytes.NewBuffer(updateBody))
	updateReq.Header.Set("Authorization", "Bearer "+authToken)
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		t.Fatalf("failed to call PUT /auth/settings: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update settings, got %d", updateResp.StatusCode)
	}

	var updateEnvelope APIEnvelope
	_ = json.NewDecoder(updateResp.Body).Decode(&updateEnvelope)
	var updatedSettings auth.UserSettings
	_ = json.Unmarshal(updateEnvelope.Data, &updatedSettings)
	if updatedSettings.CycleStartDay != 10 {
		t.Errorf("expected updated cycle_start_day 10, got %d", updatedSettings.CycleStartDay)
	}
}
