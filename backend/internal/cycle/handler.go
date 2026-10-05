package cycle

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for financial cycle endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a new cycle Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Routes mounts the cycle endpoints onto a chi.Router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/current", h.GetCurrentCycle)
	r.Get("/summary", h.GetCycleSummary)

	return r
}

func (h *Handler) GetCurrentCycle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dateParam := r.URL.Query().Get("date")
	info, err := h.service.GetCurrentCycle(r.Context(), userID, dateParam)
	if err != nil {
		if errors.Is(err, ErrInvalidDateFormat) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to calculate current cycle")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    info,
	})
}

func (h *Handler) GetCycleSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dateParam := r.URL.Query().Get("date")
	summary, err := h.service.GetCycleSummary(r.Context(), userID, dateParam)
	if err != nil {
		if errors.Is(err, ErrInvalidDateFormat) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to compute cycle summary")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    summary,
	})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]any{
		"success": false,
		"error":   message,
	})
}
