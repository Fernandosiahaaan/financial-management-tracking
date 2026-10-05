package account

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the account domain.
type Handler struct {
	service Service
}

// NewHandler creates a new account HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router with account endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateAccount)
	r.Get("/", h.ListAccounts)
	r.Get("/{id}", h.GetAccount)
	r.Put("/{id}", h.UpdateAccount)
	r.Delete("/{id}", h.DeleteAccount)

	return r
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input CreateAccountInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	acc, err := h.service.CreateAccount(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidAccountName) || errors.Is(err, ErrInvalidAccountType) || errors.Is(err, ErrNegativeOpeningBalance) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrAccountAlreadyExists) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	respondJSON(w, http.StatusCreated, acc)
}

func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.service.ListAccounts(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	respondJSON(w, http.StatusOK, accounts)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.service.GetAccount(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			respondError(w, http.StatusNotFound, "account not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get account")
		return
	}

	respondJSON(w, http.StatusOK, acc)
}

func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	var input UpdateAccountInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	acc, err := h.service.UpdateAccount(r.Context(), userID, id, input)
	if err != nil {
		if errors.Is(err, ErrInvalidAccountName) || errors.Is(err, ErrInvalidAccountType) || errors.Is(err, ErrInvalidAccountStatus) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrAccountNotFound) {
			respondError(w, http.StatusNotFound, "account not found")
			return
		}
		if errors.Is(err, ErrAccountAlreadyExists) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update account")
		return
	}

	respondJSON(w, http.StatusOK, acc)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	err = h.service.DeleteAccount(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			respondError(w, http.StatusNotFound, "account not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete account")
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
