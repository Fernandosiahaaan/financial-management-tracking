package budget

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockRepository implements Repository for testing.
type mockRepository struct {
	budgets              []Budget
	actualSpending       int64
	categoryExists       bool
	createErr            error
	listErr              error
	getErr               error
	updateErr            error
	deleteErr            error
	actualSpendingErr    error
	categoryCheckErr     error
}

func (m *mockRepository) Create(_ context.Context, b Budget) (*Budget, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	b.ID = uuid.New()
	return &b, nil
}

func (m *mockRepository) ListByUserAndCycle(_ context.Context, _ uuid.UUID, _, _ string) ([]Budget, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.budgets, nil
}

func (m *mockRepository) GetByID(_ context.Context, _, id uuid.UUID) (*Budget, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, b := range m.budgets {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, ErrBudgetNotFound
}

func (m *mockRepository) Update(_ context.Context, _, id uuid.UUID, amount int64) (*Budget, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	for _, b := range m.budgets {
		if b.ID == id {
			b.PlannedAmount = amount
			return &b, nil
		}
	}
	return nil, ErrBudgetNotFound
}

func (m *mockRepository) SoftDelete(_ context.Context, _, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for _, b := range m.budgets {
		if b.ID == id {
			return nil
		}
	}
	return ErrBudgetNotFound
}

func (m *mockRepository) GetActualSpending(_ context.Context, _, _ uuid.UUID, _, _ string) (int64, error) {
	if m.actualSpendingErr != nil {
		return 0, m.actualSpendingErr
	}
	return m.actualSpending, nil
}

func (m *mockRepository) CategoryBelongsToUser(_ context.Context, _, _ uuid.UUID) (bool, error) {
	if m.categoryCheckErr != nil {
		return false, m.categoryCheckErr
	}
	return m.categoryExists, nil
}

func TestCreateBudget_Valid(t *testing.T) {
	repo := &mockRepository{categoryExists: true, actualSpending: 150000}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CategoryID:    uuid.New(),
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 500000,
	}

	b, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if b.PlannedAmount != 500000 {
		t.Errorf("expected planned_amount 500000, got %d", b.PlannedAmount)
	}
	if b.ActualAmount != 150000 {
		t.Errorf("expected actual_amount 150000, got %d", b.ActualAmount)
	}
	if b.Variance != 350000 {
		t.Errorf("expected variance 350000, got %d", b.Variance)
	}
}

func TestCreateBudget_NegativeAmount(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CategoryID:    uuid.New(),
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: -100,
	}

	_, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != ErrInvalidPlannedAmount {
		t.Errorf("expected ErrInvalidPlannedAmount, got %v", err)
	}
}

func TestCreateBudget_InvalidCycleDates(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CategoryID:    uuid.New(),
		CycleStart:    "2026-10-31",
		CycleEnd:      "2026-10-01",
		PlannedAmount: 500000,
	}

	_, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != ErrInvalidCycleDates {
		t.Errorf("expected ErrInvalidCycleDates, got %v", err)
	}
}

func TestCreateBudget_InvalidCategoryID(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 500000,
	}

	_, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != ErrInvalidCategoryID {
		t.Errorf("expected ErrInvalidCategoryID, got %v", err)
	}
}

func TestCreateBudget_CategoryNotFound(t *testing.T) {
	repo := &mockRepository{categoryExists: false}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CategoryID:    uuid.New(),
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 500000,
	}

	_, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestListBudgets_WithVariance(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	cat1, cat2 := uuid.New(), uuid.New()

	repo := &mockRepository{
		categoryExists: true,
		actualSpending: 200000,
		budgets: []Budget{
			{ID: id1, CategoryID: cat1, CycleStart: "2026-10-01", CycleEnd: "2026-10-31", PlannedAmount: 500000},
			{ID: id2, CategoryID: cat2, CycleStart: "2026-10-01", CycleEnd: "2026-10-31", PlannedAmount: 300000},
		},
	}
	svc := NewService(repo)

	budgets, err := svc.ListBudgets(context.Background(), uuid.New(), "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(budgets) != 2 {
		t.Fatalf("expected 2 budgets, got %d", len(budgets))
	}

	// Each budget should have actual 200000 and correct variance
	for _, b := range budgets {
		if b.ActualAmount != 200000 {
			t.Errorf("expected actual_amount 200000, got %d", b.ActualAmount)
		}
		expectedVariance := b.PlannedAmount - 200000
		if b.Variance != expectedVariance {
			t.Errorf("expected variance %d, got %d", expectedVariance, b.Variance)
		}
	}
}

func TestUpdateBudget_Valid(t *testing.T) {
	budgetID := uuid.New()
	catID := uuid.New()

	repo := &mockRepository{
		categoryExists: true,
		actualSpending: 100000,
		budgets: []Budget{
			{ID: budgetID, CategoryID: catID, CycleStart: "2026-10-01", CycleEnd: "2026-10-31", PlannedAmount: 500000},
		},
	}
	svc := NewService(repo)

	b, err := svc.UpdateBudget(context.Background(), uuid.New(), budgetID, UpdateBudgetInput{PlannedAmount: 600000})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if b.PlannedAmount != 600000 {
		t.Errorf("expected planned_amount 600000, got %d", b.PlannedAmount)
	}
	if b.ActualAmount != 100000 {
		t.Errorf("expected actual_amount 100000, got %d", b.ActualAmount)
	}
	if b.Variance != 500000 {
		t.Errorf("expected variance 500000, got %d", b.Variance)
	}
}

func TestUpdateBudget_NegativeAmount(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)

	_, err := svc.UpdateBudget(context.Background(), uuid.New(), uuid.New(), UpdateBudgetInput{PlannedAmount: -1})
	if err != ErrInvalidPlannedAmount {
		t.Errorf("expected ErrInvalidPlannedAmount, got %v", err)
	}
}

func TestCreateBudget_ZeroAmount(t *testing.T) {
	repo := &mockRepository{categoryExists: true, actualSpending: 0}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CategoryID:    uuid.New(),
		CycleStart:    "2026-10-01",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 0,
	}

	b, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != nil {
		t.Fatalf("expected no error for zero planned amount, got %v", err)
	}
	if b.PlannedAmount != 0 {
		t.Errorf("expected planned_amount 0, got %d", b.PlannedAmount)
	}
}

func TestCreateBudget_InvalidDateFormat(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateBudgetInput{
		CategoryID:    uuid.New(),
		CycleStart:    "not-a-date",
		CycleEnd:      "2026-10-31",
		PlannedAmount: 500000,
	}

	_, err := svc.CreateBudget(context.Background(), uuid.New(), input)
	if err != ErrInvalidCycleDates {
		t.Errorf("expected ErrInvalidCycleDates, got %v", err)
	}
}
