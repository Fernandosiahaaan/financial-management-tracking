package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	CycleStartDay int    `json:"cycle_start_day"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateSettingsRequest struct {
	CycleStartDay int `json:"cycle_start_day"`
}

// Handler provides HTTP handlers for authentication and user preferences.
type Handler struct {
	service Service
	tokens  TokenService
}

// NewHandler creates a new auth Handler.
func NewHandler(service Service, tokens TokenService) *Handler {
	return &Handler{
		service: service,
		tokens:  tokens,
	}
}

// Routes defines public and protected authentication routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)

	// Protected routes
	r.Group(func(pr chi.Router) {
		pr.Use(RequireAuth(h.tokens))
		pr.Get("/me", h.GetMe)
		pr.Put("/settings", h.UpdateSettings)
	})

	return r
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	res, err := h.service.Register(r.Context(), req.Email, req.Password, req.CycleStartDay)
	if err != nil {
		if errors.Is(err, ErrInvalidEmail) || errors.Is(err, ErrPasswordTooShort) ||
			errors.Is(err, ErrPasswordTooLong) || errors.Is(err, ErrInvalidCycleDay) {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		if errors.Is(err, ErrUserAlreadyExists) {
			respondJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"message": "user with this email already exists",
			})
			return
		}

		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to register user",
		})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	res, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			respondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "invalid email or password",
			})
			return
		}

		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to authenticate",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    res,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// For stateless JWTs, the client discards the token.
	// This endpoint provides a symmetric API contract for clients.
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "logged out successfully",
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondUnauthorized(w, "unauthorized")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "user not found",
			})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to retrieve profile",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    profile,
	})
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		respondUnauthorized(w, "unauthorized")
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
		})
		return
	}

	settings, err := h.service.UpdateSettings(r.Context(), userID, req.CycleStartDay)
	if err != nil {
		if errors.Is(err, ErrInvalidCycleDay) {
			respondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to update settings",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    settings,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
