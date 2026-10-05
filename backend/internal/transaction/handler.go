package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for transaction endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a new transaction Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a chi.Router with transaction routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateTransaction)
	r.Get("/", h.ListTransactions)
	r.Get("/{id}", h.GetTransaction)
	r.Put("/{id}", h.UpdateTransaction)
	r.Delete("/{id}", h.DeleteTransaction)

	return r
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.service.CreateTransaction(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType),
			errors.Is(err, ErrInvalidAmount),
			errors.Is(err, ErrInvalidDate),
			errors.Is(err, ErrMissingAccount),
			errors.Is(err, ErrMissingDestAccount),
			errors.Is(err, ErrSameAccountTransfer),
			errors.Is(err, ErrTransferDestinationSet):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrAccountNotFound),
			errors.Is(err, ErrCategoryNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to create transaction")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    tx,
		"message": "transaction recorded successfully",
	})
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	filter := ListFilter{
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
	}

	if t := q.Get("type"); t != "" {
		txType := Type(t)
		filter.Type = &txType
	}

	if acc := q.Get("account_id"); acc != "" {
		if accID, err := uuid.Parse(acc); err == nil {
			filter.AccountID = &accID
		}
	}

	if lim := q.Get("limit"); lim != "" {
		if l, err := strconv.Atoi(lim); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if off := q.Get("offset"); off != "" {
		if o, err := strconv.Atoi(off); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	transactions, err := h.service.ListTransactions(r.Context(), userID, filter)
	if err != nil {
		if errors.Is(err, ErrInvalidDate) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    transactions,
	})
}

func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id format")
		return
	}

	tx, err := h.service.GetTransaction(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			respondError(w, http.StatusNotFound, "transaction not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get transaction")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tx,
	})
}

func (h *Handler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id format")
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.service.UpdateTransaction(r.Context(), userID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType),
			errors.Is(err, ErrInvalidAmount),
			errors.Is(err, ErrInvalidDate),
			errors.Is(err, ErrMissingAccount),
			errors.Is(err, ErrMissingDestAccount),
			errors.Is(err, ErrSameAccountTransfer),
			errors.Is(err, ErrTransferDestinationSet):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrTransactionNotFound),
			errors.Is(err, ErrAccountNotFound):
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update transaction")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tx,
		"message": "transaction updated successfully",
	})
}

func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid transaction id format")
		return
	}

	if err := h.service.DeleteTransaction(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			respondError(w, http.StatusNotFound, "transaction not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
