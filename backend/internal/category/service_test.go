package category

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockCategoryRepo struct {
	categories map[uuid.UUID]Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[uuid.UUID]Category)}
}

func (m *mockCategoryRepo) Create(ctx context.Context, cat Category) (*Category, error) {
	for _, c := range m.categories {
		if c.UserID == cat.UserID && strings.EqualFold(c.Name, cat.Name) && c.Type == cat.Type && c.DeletedAt == nil {
			return nil, ErrCategoryAlreadyExists
		}
	}
	cat.ID = uuid.New()
	cat.CreatedAt = time.Now()
	cat.UpdatedAt = time.Now()
	m.categories[cat.ID] = cat
	return &cat, nil
}

func (m *mockCategoryRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	var list []Category
	for _, c := range m.categories {
		if c.UserID == userID && c.DeletedAt == nil {
			list = append(list, c)
		}
	}
	return list, nil
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*Category, error) {
	c, ok := m.categories[id]
	if !ok || c.UserID != userID || c.DeletedAt != nil {
		return nil, ErrCategoryNotFound
	}
	return &c, nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, cat Category) (*Category, error) {
	existing, ok := m.categories[cat.ID]
	if !ok || existing.UserID != cat.UserID || existing.DeletedAt != nil {
		return nil, ErrCategoryNotFound
	}
	for _, c := range m.categories {
		if c.ID != cat.ID && c.UserID == cat.UserID && strings.EqualFold(c.Name, cat.Name) && c.Type == cat.Type && c.DeletedAt == nil {
			return nil, ErrCategoryAlreadyExists
		}
	}
	existing.Name = cat.Name
	existing.Type = cat.Type
	existing.Icon = cat.Icon
	existing.Color = cat.Color
	existing.UpdatedAt = time.Now()
	m.categories[cat.ID] = existing
	return &existing, nil
}

func (m *mockCategoryRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	existing, ok := m.categories[id]
	if !ok || existing.UserID != userID || existing.DeletedAt != nil {
		return ErrCategoryNotFound
	}
	now := time.Now()
	existing.DeletedAt = &now
	m.categories[id] = existing
	return nil
}

func TestCategoryService_Validation(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	userID := uuid.New()

	// Empty name
	_, err := svc.CreateCategory(context.Background(), userID, CreateCategoryInput{
		Name: "",
		Type: TypeExpense,
	})
	if err != ErrInvalidCategoryName {
		t.Fatalf("expected ErrInvalidCategoryName, got %v", err)
	}

	// Invalid type
	_, err = svc.CreateCategory(context.Background(), userID, CreateCategoryInput{
		Name: "Food",
		Type: Type("SAVINGS"),
	})
	if err != ErrInvalidCategoryType {
		t.Fatalf("expected ErrInvalidCategoryType, got %v", err)
	}

	// Success with defaults
	cat, err := svc.CreateCategory(context.Background(), userID, CreateCategoryInput{
		Name: "Food & Dining",
		Type: TypeExpense,
	})
	if err != nil {
		t.Fatalf("unexpected error creating category: %v", err)
	}
	if cat.Icon != "tag" {
		t.Errorf("expected default icon 'tag', got '%s'", cat.Icon)
	}
	if cat.Color != "#6366F1" {
		t.Errorf("expected default color '#6366F1', got '%s'", cat.Color)
	}

	// Duplicate
	_, err = svc.CreateCategory(context.Background(), userID, CreateCategoryInput{
		Name: "food & dining",
		Type: TypeExpense,
	})
	if err != ErrCategoryAlreadyExists {
		t.Fatalf("expected ErrCategoryAlreadyExists, got %v", err)
	}

	// Same name but different type (INCOME) allowed
	_, err = svc.CreateCategory(context.Background(), userID, CreateCategoryInput{
		Name: "Food & Dining",
		Type: TypeIncome,
	})
	if err != nil {
		t.Fatalf("expected different type to succeed, got %v", err)
	}
}

func TestCategoryService_UpdateAndDelete(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	userID := uuid.New()

	cat, _ := svc.CreateCategory(context.Background(), userID, CreateCategoryInput{
		Name:  "Salary",
		Type:  TypeIncome,
		Icon:  "briefcase",
		Color: "#10B981",
	})

	updated, err := svc.UpdateCategory(context.Background(), userID, cat.ID, UpdateCategoryInput{
		Name:  "Primary Salary",
		Type:  TypeIncome,
		Icon:  "wallet",
		Color: "#10B981",
	})
	if err != nil {
		t.Fatalf("unexpected error updating category: %v", err)
	}
	if updated.Name != "Primary Salary" || updated.Icon != "wallet" {
		t.Errorf("unexpected updated values: %+v", updated)
	}

	// Delete
	err = svc.DeleteCategory(context.Background(), userID, cat.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting category: %v", err)
	}

	// Verify not found
	_, err = svc.GetCategory(context.Background(), userID, cat.ID)
	if err != ErrCategoryNotFound {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}
