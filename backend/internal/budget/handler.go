package budget

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for budgets.
type Handler struct {
	service Service
}

// NewHandler creates a new budget HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router with budget endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateBudget)
	r.Get("/", h.ListBudgets)
	r.Get("/{id}", h.GetBudget)
	r.Put("/{id}", h.UpdateBudget)
	r.Delete("/{id}", h.DeleteBudget)

	return r
}

func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input CreateBudgetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.service.CreateBudget(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidPlannedAmount) || errors.Is(err, ErrInvalidCycleDates) || errors.Is(err, ErrInvalidCategoryID) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrCategoryNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrBudgetAlreadyExists) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create budget")
		return
	}

	respondJSON(w, http.StatusCreated, b)
}

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cycleStart := r.URL.Query().Get("cycle_start")
	cycleEnd := r.URL.Query().Get("cycle_end")

	if cycleStart == "" || cycleEnd == "" {
		respondError(w, http.StatusBadRequest, "cycle_start and cycle_end query parameters are required")
		return
	}

	budgets, err := h.service.ListBudgets(r.Context(), userID, cycleStart, cycleEnd)
	if err != nil {
		if errors.Is(err, ErrInvalidCycleDates) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to list budgets")
		return
	}

	respondJSON(w, http.StatusOK, budgets)
}

func (h *Handler) GetBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid budget ID")
		return
	}

	b, err := h.service.GetBudget(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrBudgetNotFound) {
			respondError(w, http.StatusNotFound, "budget not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get budget")
		return
	}

	respondJSON(w, http.StatusOK, b)
}

func (h *Handler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid budget ID")
		return
	}

	var input UpdateBudgetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.service.UpdateBudget(r.Context(), userID, id, input)
	if err != nil {
		if errors.Is(err, ErrInvalidPlannedAmount) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrBudgetNotFound) {
			respondError(w, http.StatusNotFound, "budget not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update budget")
		return
	}

	respondJSON(w, http.StatusOK, b)
}

func (h *Handler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid budget ID")
		return
	}

	err = h.service.DeleteBudget(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrBudgetNotFound) {
			respondError(w, http.StatusNotFound, "budget not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete budget")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
