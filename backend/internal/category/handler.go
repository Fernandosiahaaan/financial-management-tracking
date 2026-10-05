package category

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for categories.
type Handler struct {
	service Service
}

// NewHandler creates a new category HTTP Handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a Chi router with category endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateCategory)
	r.Get("/", h.ListCategories)
	r.Get("/{id}", h.GetCategory)
	r.Put("/{id}", h.UpdateCategory)
	r.Delete("/{id}", h.DeleteCategory)

	return r
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input CreateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.service.CreateCategory(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidCategoryName) || errors.Is(err, ErrInvalidCategoryType) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrCategoryAlreadyExists) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	respondJSON(w, http.StatusCreated, cat)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	categories, err := h.service.ListCategories(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	cat, err := h.service.GetCategory(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			respondError(w, http.StatusNotFound, "category not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get category")
		return
	}

	respondJSON(w, http.StatusOK, cat)
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var input UpdateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.service.UpdateCategory(r.Context(), userID, id, input)
	if err != nil {
		if errors.Is(err, ErrInvalidCategoryName) || errors.Is(err, ErrInvalidCategoryType) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrCategoryNotFound) {
			respondError(w, http.StatusNotFound, "category not found")
			return
		}
		if errors.Is(err, ErrCategoryAlreadyExists) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update category")
		return
	}

	respondJSON(w, http.StatusOK, cat)
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	err = h.service.DeleteCategory(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			respondError(w, http.StatusNotFound, "category not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete category")
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
