package category

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidCategoryName = errors.New("category name must be between 1 and 100 characters")
	ErrInvalidCategoryType = errors.New("category type must be either INCOME or EXPENSE")
)

type CreateCategoryInput struct {
	Name  string `json:"name"`
	Type  Type   `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type UpdateCategoryInput struct {
	Name  string `json:"name"`
	Type  Type   `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// Service defines category business operations.
type Service interface {
	CreateCategory(ctx context.Context, userID uuid.UUID, input CreateCategoryInput) (*Category, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetCategory(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	UpdateCategory(ctx context.Context, userID, id uuid.UUID, input UpdateCategoryInput) (*Category, error)
	DeleteCategory(ctx context.Context, userID, id uuid.UUID) error
}

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// NewService creates a new category Service.
func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

func (s *DefaultService) CreateCategory(ctx context.Context, userID uuid.UUID, input CreateCategoryInput) (*Category, error) {
	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" || len(cleanName) > 100 {
		return nil, ErrInvalidCategoryName
	}
	if !input.Type.Valid() {
		return nil, ErrInvalidCategoryType
	}

	icon := strings.TrimSpace(input.Icon)
	if icon == "" {
		icon = "tag"
	}
	color := strings.TrimSpace(input.Color)
	if color == "" {
		color = "#6366F1"
	}

	cat := Category{
		UserID: userID,
		Name:   cleanName,
		Type:   input.Type,
		Icon:   icon,
		Color:  color,
	}

	return s.repo.Create(ctx, cat)
}

func (s *DefaultService) ListCategories(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *DefaultService) GetCategory(ctx context.Context, userID, id uuid.UUID) (*Category, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *DefaultService) UpdateCategory(ctx context.Context, userID, id uuid.UUID, input UpdateCategoryInput) (*Category, error) {
	cleanName := strings.TrimSpace(input.Name)
	if cleanName == "" || len(cleanName) > 100 {
		return nil, ErrInvalidCategoryName
	}
	if !input.Type.Valid() {
		return nil, ErrInvalidCategoryType
	}

	icon := strings.TrimSpace(input.Icon)
	if icon == "" {
		icon = "tag"
	}
	color := strings.TrimSpace(input.Color)
	if color == "" {
		color = "#6366F1"
	}

	cat := Category{
		ID:     id,
		UserID: userID,
		Name:   cleanName,
		Type:   input.Type,
		Icon:   icon,
		Color:  color,
	}

	return s.repo.Update(ctx, cat)
}

func (s *DefaultService) DeleteCategory(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, userID, id)
}
