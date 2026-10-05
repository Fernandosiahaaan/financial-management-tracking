package allocation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for allocations.
type Handler struct {
	service Service
}

// NewHandler creates a new allocation HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router with allocation endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateAllocation)
	r.Get("/", h.ListAllocations)
	r.Get("/{id}", h.GetAllocation)
	r.Put("/{id}", h.UpdateAllocation)
	r.Delete("/{id}", h.DeleteAllocation)

	return r
}

func (h *Handler) CreateAllocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input CreateAllocationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	a, err := h.service.CreateAllocation(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidAllocatedAmount) || errors.Is(err, ErrInvalidCycleDates) || errors.Is(err, ErrInvalidCategoryID) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrCategoryNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrAllocationAlreadyExists) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, ErrExceedsAvailableIncome) {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create allocation")
		return
	}

	respondJSON(w, http.StatusCreated, a)
}

func (h *Handler) ListAllocations(w http.ResponseWriter, r *http.Request) {
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

	summary, err := h.service.ListAllocations(r.Context(), userID, cycleStart, cycleEnd)
	if err != nil {
		if errors.Is(err, ErrInvalidCycleDates) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to list allocations")
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (h *Handler) GetAllocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid allocation ID")
		return
	}

	a, err := h.service.GetAllocation(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrAllocationNotFound) {
			respondError(w, http.StatusNotFound, "allocation not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get allocation")
		return
	}

	respondJSON(w, http.StatusOK, a)
}

func (h *Handler) UpdateAllocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid allocation ID")
		return
	}

	var input UpdateAllocationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	a, err := h.service.UpdateAllocation(r.Context(), userID, id, input)
	if err != nil {
		if errors.Is(err, ErrInvalidAllocatedAmount) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrAllocationNotFound) {
			respondError(w, http.StatusNotFound, "allocation not found")
			return
		}
		if errors.Is(err, ErrExceedsAvailableIncome) {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update allocation")
		return
	}

	respondJSON(w, http.StatusOK, a)
}

func (h *Handler) DeleteAllocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid allocation ID")
		return
	}

	err = h.service.DeleteAllocation(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrAllocationNotFound) {
			respondError(w, http.StatusNotFound, "allocation not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete allocation")
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
