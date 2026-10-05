package report

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for financial reports and asset snapshots.
type Handler struct {
	service Service
}

// NewHandler creates a new report HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router for reports.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/monthly", h.GetMonthlyReport)
	r.Get("/cycle", h.GetCycleReport)
	return r
}

// GetMonthlyReport returns a calendar-month financial report.
func (h *Handler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	now := time.Now().UTC()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		parsedYear, err := strconv.Atoi(yearStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid year parameter")
			return
		}
		year = parsedYear
	}

	if monthStr != "" {
		parsedMonth, err := strconv.Atoi(monthStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid month parameter")
			return
		}
		month = parsedMonth
	}

	report, err := h.service.GenerateMonthlyReport(r.Context(), userID, year, month)
	if err != nil {
		if errors.Is(err, ErrInvalidYear) || errors.Is(err, ErrInvalidMonth) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to generate monthly report")
		return
	}

	respondJSON(w, http.StatusOK, report)
}

// GetCycleReport returns a financial-cycle financial report.
func (h *Handler) GetCycleReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dateParam := r.URL.Query().Get("date")

	report, err := h.service.GenerateCycleReport(r.Context(), userID, dateParam)
	if err != nil {
		if errors.Is(err, ErrInvalidDateFormat) {
			respondError(w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to generate cycle report")
		return
	}

	respondJSON(w, http.StatusOK, report)
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
