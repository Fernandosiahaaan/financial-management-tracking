package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRequireAuth_Success(t *testing.T) {
	tokens := NewJWTService("test-secret-key-1234567890123456", 1*time.Hour)
	userID := uuid.New()
	user := User{ID: userID, Email: "auth@example.com"}

	tokenStr, err := tokens.GenerateToken(user)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	var extractedID uuid.UUID
	var called bool

	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		id, ok := GetUserID(r.Context())
		if !ok {
			t.Fatal("expected user ID in context")
		}
		extractedID = id
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if extractedID != userID {
		t.Fatalf("expected extracted ID %v, got %v", userID, extractedID)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	tokens := NewJWTService("test-secret-key-1234567890123456", 1*time.Hour)

	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	tokens := NewJWTService("test-secret-key-1234567890123456", 1*time.Hour)

	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	tokens := NewJWTService("test-secret-key-1234567890123456", 1*time.Hour)

	handler := RequireAuth(tokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
