package account

import (
	"context"
	"errors"
	"strings"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
	"github.com/google/uuid"
)

var (
	ErrInvalidAccountName     = errors.New("account name must be between 1 and 100 characters")
	ErrInvalidAccountType     = errors.New("invalid account type")
	ErrInvalidAccountStatus   = errors.New("invalid account status")
	ErrNegativeOpeningBalance = errors.New("opening balance cannot be negative")
)

type CreateAccountInput struct {
	Name           string      `json:"name"`
	Type           Type        `json:"type"`
	OpeningBalance money.Cents `json:"opening_balance"`
}

type UpdateAccountInput struct {
	Name   string `json:"name"`
	Type   Type   `json:"type"`
	Status Status `json:"status"`
}

// Service defines account business logic operations.
type Service interface {
	CreateAccount(ctx context.Context, userID uuid.UUID, input CreateAccountInput) (*Account, error)
	ListAccounts(ctx context.Context, userID uuid.UUID) ([]Account, error)
	GetAccount(ctx context.Context, userID, id uuid.UUID) (*Account, error)
	UpdateAccount(ctx context.Context, userID, id uuid.UUID, input UpdateAccountInput) (*Account, error)
	DeleteAccount(ctx context.Context, userID, id uuid.UUID) error
}

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// NewService creates a new account Service.
func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

func (s *DefaultService) CreateAccount(ctx context.Context, userID uuid.UUID, input CreateAccountInput) (*Account, error) {
	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" || len(cleanName) > 100 {
		return nil, ErrInvalidAccountName
	}
	if !input.Type.Valid() {
		return nil, ErrInvalidAccountType
	}
	if input.OpeningBalance < 0 {
		return nil, ErrNegativeOpeningBalance
	}

	acc := Account{
		UserID:         userID,
		Name:           cleanName,
		Type:           input.Type,
		OpeningBalance: input.OpeningBalance.Int64(),
		CurrentBalance: input.OpeningBalance.Int64(),
		Status:         StatusActive,
	}

	return s.repo.Create(ctx, acc)
}

func (s *DefaultService) ListAccounts(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *DefaultService) GetAccount(ctx context.Context, userID, id uuid.UUID) (*Account, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *DefaultService) UpdateAccount(ctx context.Context, userID, id uuid.UUID, input UpdateAccountInput) (*Account, error) {
	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" || len(cleanName) > 100 {
		return nil, ErrInvalidAccountName
	}
	if !input.Type.Valid() {
		return nil, ErrInvalidAccountType
	}
	if input.Status != "" && !input.Status.Valid() {
		return nil, ErrInvalidAccountStatus
	}
	if input.Status == "" {
		input.Status = StatusActive
	}

	acc := Account{
		ID:     id,
		UserID: userID,
		Name:   cleanName,
		Type:   input.Type,
		Status: input.Status,
	}

	return s.repo.Update(ctx, acc)
}

func (s *DefaultService) DeleteAccount(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, userID, id)
}
