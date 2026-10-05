package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernando/financial-management-tracking/backend/internal/middleware"
)

func TestRequestID(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: Generates ID
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}

	// Case 2: Preserves existing ID
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Request-ID", "custom-id-123")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Header().Get("X-Request-ID") != "custom-id-123" {
		t.Fatalf("expected custom-id-123, got %s", rec2.Header().Get("X-Request-ID"))
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for k, expected := range headers {
		if val := rec.Header().Get(k); val != expected {
			t.Errorf("header %s: expected %s, got %s", k, expected, val)
		}
	}
}

func TestMaxBodySize(t *testing.T) {
	limit := int64(10) // 10 bytes limit
	handler := middleware.MaxBodySize(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: Below limit
	reqSmall := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("small"))
	recSmall := httptest.NewRecorder()
	handler.ServeHTTP(recSmall, reqSmall)
	if recSmall.Code != http.StatusOK {
		t.Errorf("expected 200 OK for small body, got %d", recSmall.Code)
	}

	// Case 2: Above limit
	reqLarge := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("this payload is definitely larger than ten bytes"))
	recLarge := httptest.NewRecorder()
	handler.ServeHTTP(recLarge, reqLarge)
	if recLarge.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for oversized body, got %d", recLarge.Code)
	}
}

func TestRateLimiter(t *testing.T) {
	// 60 requests per minute with burst of 2
	limiter := middleware.NewRateLimiter(60, 2)
	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "192.168.1.100:1234"

	// Request 1: Allowed (burst 1 used)
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	req1.RemoteAddr = ip
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected request 1 to be 200, got %d", rec1.Code)
	}

	// Request 2: Allowed (burst 2 used)
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	req2.RemoteAddr = ip
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected request 2 to be 200, got %d", rec2.Code)
	}

	// Request 3: Blocked (rate limit exceeded)
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	req3.RemoteAddr = ip
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected request 3 to be 429 Too Many Requests, got %d", rec3.Code)
	}
	if rec3.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After header 60, got %s", rec3.Header().Get("Retry-After"))
	}

	// Request 4: Health endpoint bypasses rate limiting
	reqHealth := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	reqHealth.RemoteAddr = ip
	recHealth := httptest.NewRecorder()
	handler.ServeHTTP(recHealth, reqHealth)
	if recHealth.Code != http.StatusOK {
		t.Fatalf("expected /health to bypass rate limit with 200, got %d", recHealth.Code)
	}
}

func TestRecoverer(t *testing.T) {
	handler := middleware.Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected critical failure")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on panic recovery, got %d", rec.Code)
	}
}

func TestCORS(t *testing.T) {
	handler := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Preflight OPTIONS
	reqOpt := httptest.NewRequest(http.MethodOptions, "/test", nil)
	recOpt := httptest.NewRecorder()
	handler.ServeHTTP(recOpt, reqOpt)

	if recOpt.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", recOpt.Code)
	}
	if recOpt.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS origin header, got %s", recOpt.Header().Get("Access-Control-Allow-Origin"))
	}
}
