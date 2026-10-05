package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/fernando/financial-management-tracking/backend/internal/cycle"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for dashboard metrics.
type Handler struct {
	service Service
}

// NewHandler creates a new dashboard HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router with the dashboard endpoint.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetDashboard)
	return r
}

// GetDashboard returns the full dashboard metrics for the authenticated user.
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dateParam := r.URL.Query().Get("date")

	data, err := h.service.GetDashboardData(r.Context(), userID, dateParam)
	if err != nil {
		if errors.Is(err, cycle.ErrInvalidDateFormat) {
			respondError(w, http.StatusBadRequest, "invalid date format, must be YYYY-MM-DD")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to load dashboard data")
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
