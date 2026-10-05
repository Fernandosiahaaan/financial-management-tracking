package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestHandler_Healthy(t *testing.T) {
	mock := &mockPinger{err: nil}
	handler := Handler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", rec.Code)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got %v", resp.Success)
	}
	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", resp.Status)
	}
	if resp.Database != "connected" {
		t.Errorf("expected database 'connected', got '%s'", resp.Database)
	}
}

func TestHandler_Unhealthy(t *testing.T) {
	mock := &mockPinger{err: errors.New("connection refused")}
	handler := Handler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status code 503, got %d", rec.Code)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success false, got %v", resp.Success)
	}
	if resp.Status != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got '%s'", resp.Status)
	}
	if resp.Database != "disconnected" {
		t.Errorf("expected database 'disconnected', got '%s'", resp.Database)
	}
}

func TestHandler_NilPinger(t *testing.T) {
	handler := Handler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status code 503, got %d", rec.Code)
	}

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success false, got %v", resp.Success)
	}
	if resp.Status != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got '%s'", resp.Status)
	}
	if resp.Database != "disconnected" {
		t.Errorf("expected database 'disconnected', got '%s'", resp.Database)
	}
}

func TestResponse_JSONStructure(t *testing.T) {
	resp := Response{
		Success:  true,
		Status:   "healthy",
		Database: "connected",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !decoded.Success {
		t.Errorf("expected success=true, got %v", decoded.Success)
	}
	if decoded.Status != "healthy" {
		t.Errorf("expected status=healthy, got %s", decoded.Status)
	}
	if decoded.Database != "connected" {
		t.Errorf("expected database=connected, got %s", decoded.Database)
	}
}
