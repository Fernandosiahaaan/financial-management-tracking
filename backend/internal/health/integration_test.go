package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/fernando/financial-management-tracking/backend/internal/health"
	"github.com/go-chi/chi/v5"
)

func TestHealthEndpoint_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback to local default if running in dev environment
		dsn = "postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping integration test: database not reachable at %s (%v)", dsn, err)
		return
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Get("/api/v1/health", health.Handler(pool))

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. With reachable DB, expect 200 OK
	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("failed to GET health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// 2. Close pool to simulate DB disconnection, expect 503
	pool.Close()

	resp2, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("failed to GET health after close: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 after DB pool closed, got %d", resp2.StatusCode)
	}
}
