package receivable

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the receivables domain.
type Handler struct {
	service *Service
}

// NewHandler creates a new receivable HTTP Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router with receivable endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Post("/{id}/payments", h.RecordPayment)
	r.Patch("/{id}", h.UpdateStatus)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateReceivableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rec, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidCounterparty) ||
			errors.Is(err, ErrInvalidPrincipal) ||
			errors.Is(err, ErrInvalidDueDate) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrAccountNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrInsufficientBalance) {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create receivable: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, rec)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var filter ListFilter
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		st := Status(statusParam)
		filter.Status = &st
	}

	receivables, err := h.service.List(r.Context(), userID, filter)
	if err != nil {
		if errors.Is(err, ErrInvalidStatus) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to list receivables: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, receivables)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid receivable id")
		return
	}

	rec, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrReceivableNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get receivable: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rec)
}

func (h *Handler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	receivableID, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid receivable id")
		return
	}

	var req RecordPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payment, updatedRec, err := h.service.RecordPayment(r.Context(), userID, receivableID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidPaymentAmount) ||
			errors.Is(err, ErrInvalidPaymentDate) ||
			errors.Is(err, ErrMissingTargetAccount) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrReceivableNotFound) || errors.Is(err, ErrAccountNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrOverpayment) || errors.Is(err, ErrAlreadySettled) {
			respondError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to record payment: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"payment":    payment,
		"receivable": updatedRec,
	})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid receivable id")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rec, err := h.service.UpdateStatus(r.Context(), userID, id, req)
	if err != nil {
		if errors.Is(err, ErrInvalidStatus) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrReceivableNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update status: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, rec)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid receivable id")
		return
	}

	if err := h.service.SoftDelete(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrReceivableNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete receivable: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "receivable deleted successfully",
	})
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
