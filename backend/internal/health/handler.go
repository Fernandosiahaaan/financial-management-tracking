package health

import (
	"context"
	"encoding/json"
	"net/http"
)

// Pinger defines the contract for checking database connectivity.
// Both *pgxpool.Pool and test mocks satisfy this interface.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Response is the JSON payload returned by the health endpoint.
type Response struct {
	Success  bool   `json:"success"`
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Handler returns an http.HandlerFunc that reports application health.
// It checks database connectivity and responds with 200 OK when healthy,
// or 503 Service Unavailable when the database is unreachable or nil.
func Handler(pinger Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dbStatus := "connected"
		httpStatus := http.StatusOK
		appStatus := "healthy"

		if pinger == nil || pinger.Ping(r.Context()) != nil {
			dbStatus = "disconnected"
			httpStatus = http.StatusServiceUnavailable
			appStatus = "unhealthy"
		}

		resp := Response{
			Success:  httpStatus == http.StatusOK,
			Status:   appStatus,
			Database: dbStatus,
		}

		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
