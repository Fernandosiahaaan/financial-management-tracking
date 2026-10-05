package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail       = errors.New("invalid email address format")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidCycleDay    = errors.New("cycle start day must be between 1 and 31")
)

// AuthResponse is returned on successful registration or login.
type AuthResponse struct {
	Token   string      `json:"token"`
	Profile UserProfile `json:"profile"`
}

// Service defines authentication and user preference operations.
type Service interface {
	Register(ctx context.Context, email, password string, cycleStartDay int) (*AuthResponse, error)
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	UpdateSettings(ctx context.Context, userID uuid.UUID, cycleStartDay int) (*UserSettings, error)
}

// DefaultService implements Service.
type DefaultService struct {
	repo   Repository
	tokens TokenService
}

// NewService creates a new DefaultService.
func NewService(repo Repository, tokens TokenService) *DefaultService {
	return &DefaultService{
		repo:   repo,
		tokens: tokens,
	}
}

// Register registers a new user with email, hashed password, and initializes settings.
func (s *DefaultService) Register(ctx context.Context, email, password string, cycleStartDay int) (*AuthResponse, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return nil, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(cleanEmail)
	if err != nil || addr.Address != cleanEmail {
		return nil, ErrInvalidEmail
	}

	if cycleStartDay <= 0 {
		cycleStartDay = 1
	}
	if cycleStartDay < 1 || cycleStartDay > 31 {
		return nil, ErrInvalidCycleDay
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := User{
		Email:        cleanEmail,
		PasswordHash: hash,
	}

	created, err := s.repo.CreateUser(ctx, user, cycleStartDay)
	if err != nil {
		return nil, err
	}

	token, err := s.tokens.GenerateToken(*created)
	if err != nil {
		return nil, fmt.Errorf("register: generate token: %w", err)
	}

	profile := UserProfile{
		ID:            created.ID,
		Email:         created.Email,
		CycleStartDay: cycleStartDay,
		CreatedAt:     created.CreatedAt,
	}

	return &AuthResponse{
		Token:   token,
		Profile: profile,
	}, nil
}

// Login verifies user credentials and returns an auth token upon success.
func (s *DefaultService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: get user: %w", err)
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.tokens.GenerateToken(*user)
	if err != nil {
		return nil, fmt.Errorf("login: generate token: %w", err)
	}

	profile, err := s.repo.GetUserProfile(ctx, user.ID)
	if err != nil {
		// Fallback profile if settings lookup fails
		profile = &UserProfile{
			ID:            user.ID,
			Email:         user.Email,
			CycleStartDay: 1,
			CreatedAt:     user.CreatedAt,
		}
	}

	return &AuthResponse{
		Token:   token,
		Profile: *profile,
	}, nil
}

// GetProfile retrieves the profile and preferences for the given user ID.
func (s *DefaultService) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	return s.repo.GetUserProfile(ctx, userID)
}

// UpdateSettings updates user preferences like cycle_start_day.
func (s *DefaultService) UpdateSettings(ctx context.Context, userID uuid.UUID, cycleStartDay int) (*UserSettings, error) {
	if cycleStartDay < 1 || cycleStartDay > 31 {
		return nil, ErrInvalidCycleDay
	}
	return s.repo.UpdateUserSettings(ctx, userID, cycleStartDay)
}
