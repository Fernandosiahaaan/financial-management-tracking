package allocation

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// mockRepository implements Repository for testing.
type mockRepository struct {
	allocations      []Allocation
	totalAllocated   int64
	cycleIncome      int64
	categoryExists   bool
	createErr        error
	listErr          error
	getErr           error
	updateErr        error
	deleteErr        error
	totalErr         error
	incomeErr        error
	categoryCheckErr error
}

func (m *mockRepository) Create(_ context.Context, a Allocation) (*Allocation, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	a.ID = uuid.New()
	return &a, nil
}

func (m *mockRepository) ListByUserAndCycle(_ context.Context, _ uuid.UUID, _, _ string) ([]Allocation, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.allocations, nil
}

func (m *mockRepository) GetByID(_ context.Context, _, id uuid.UUID) (*Allocation, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, a := range m.allocations {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, ErrAllocationNotFound
}

func (m *mockRepository) Update(_ context.Context, _, id uuid.UUID, amount int64) (*Allocation, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	for _, a := range m.allocations {
		if a.ID == id {
			a.AllocatedAmount = amount
			return &a, nil
		}
	}
	return nil, ErrAllocationNotFound
}

func (m *mockRepository) SoftDelete(_ context.Context, _, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for _, a := range m.allocations {
		if a.ID == id {
			return nil
		}
	}
	return ErrAllocationNotFound
}

func (m *mockRepository) GetTotalAllocated(_ context.Context, _ uuid.UUID, _, _ string, _ *uuid.UUID) (int64, error) {
	if m.totalErr != nil {
		return 0, m.totalErr
	}
	return m.totalAllocated, nil
}

func (m *mockRepository) GetCycleIncome(_ context.Context, _ uuid.UUID, _, _ string) (int64, error) {
	if m.incomeErr != nil {
		return 0, m.incomeErr
	}
	return m.cycleIncome, nil
}

func (m *mockRepository) CategoryBelongsToUser(_ context.Context, _, _ uuid.UUID) (bool, error) {
	if m.categoryCheckErr != nil {
		return false, m.categoryCheckErr
	}
	return m.categoryExists, nil
}

func TestCreateAllocation_Valid(t *testing.T) {
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    10000000,
		totalAllocated: 3000000,
	}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 5000000,
	}

	a, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.AllocatedAmount != 5000000 {
		t.Errorf("expected allocated_amount 5000000, got %d", a.AllocatedAmount)
	}
}

func TestCreateAllocation_ExceedsIncome(t *testing.T) {
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    10000000,
		totalAllocated: 8000000,
	}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 5000000, // 8M + 5M = 13M > 10M
	}

	_, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != ErrExceedsAvailableIncome {
		t.Errorf("expected ErrExceedsAvailableIncome, got %v", err)
	}
}

func TestCreateAllocation_NegativeAmount(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: -100,
	}

	_, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != ErrInvalidAllocatedAmount {
		t.Errorf("expected ErrInvalidAllocatedAmount, got %v", err)
	}
}

func TestCreateAllocation_InvalidCategoryID(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 1000000,
	}

	_, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != ErrInvalidCategoryID {
		t.Errorf("expected ErrInvalidCategoryID, got %v", err)
	}
}

func TestCreateAllocation_CategoryNotFound(t *testing.T) {
	repo := &mockRepository{categoryExists: false}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 1000000,
	}

	_, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCreateAllocation_InvalidCycleDates(t *testing.T) {
	repo := &mockRepository{categoryExists: true}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-31",
		CycleEnd:        "2026-10-01",
		AllocatedAmount: 1000000,
	}

	_, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != ErrInvalidCycleDates {
		t.Errorf("expected ErrInvalidCycleDates, got %v", err)
	}
}

func TestCreateAllocation_ExactlyMatchesIncome(t *testing.T) {
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    10000000,
		totalAllocated: 5000000,
	}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 5000000, // 5M + 5M = 10M == 10M (exactly)
	}

	a, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != nil {
		t.Fatalf("expected no error for exact income match, got %v", err)
	}
	if a.AllocatedAmount != 5000000 {
		t.Errorf("expected 5000000, got %d", a.AllocatedAmount)
	}
}

func TestListAllocations_Summary(t *testing.T) {
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    10000000,
		totalAllocated: 7000000,
		allocations: []Allocation{
			{ID: uuid.New(), AllocatedAmount: 4000000},
			{ID: uuid.New(), AllocatedAmount: 3000000},
		},
	}
	svc := NewService(repo)

	summary, err := svc.ListAllocations(context.Background(), uuid.New(), "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.TotalAllocated != 7000000 {
		t.Errorf("expected total_allocated 7000000, got %d", summary.TotalAllocated)
	}
	if summary.TotalIncome != 10000000 {
		t.Errorf("expected total_income 10000000, got %d", summary.TotalIncome)
	}
	if summary.RemainingIncome != 3000000 {
		t.Errorf("expected remaining_income 3000000, got %d", summary.RemainingIncome)
	}
	if len(summary.Allocations) != 2 {
		t.Errorf("expected 2 allocations, got %d", len(summary.Allocations))
	}
}

func TestUpdateAllocation_ExceedsIncome(t *testing.T) {
	allocID := uuid.New()
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    10000000,
		totalAllocated: 8000000, // other allocations sum
		allocations: []Allocation{
			{ID: allocID, CycleStart: "2026-10-01", CycleEnd: "2026-10-31", AllocatedAmount: 2000000},
		},
	}
	svc := NewService(repo)

	_, err := svc.UpdateAllocation(context.Background(), uuid.New(), allocID, UpdateAllocationInput{AllocatedAmount: 5000000})
	// 8M (others) + 5M (new) = 13M > 10M
	if err != ErrExceedsAvailableIncome {
		t.Errorf("expected ErrExceedsAvailableIncome, got %v", err)
	}
}

func TestCreateAllocation_ZeroIncome(t *testing.T) {
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    0,
		totalAllocated: 0,
	}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 1000000,
	}

	_, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != ErrExceedsAvailableIncome {
		t.Errorf("expected ErrExceedsAvailableIncome when income is 0, got %v", err)
	}
}

func TestCreateAllocation_ZeroAmountWithZeroIncome(t *testing.T) {
	repo := &mockRepository{
		categoryExists: true,
		cycleIncome:    0,
		totalAllocated: 0,
	}
	svc := NewService(repo)

	input := CreateAllocationInput{
		CategoryID:      uuid.New(),
		CycleStart:      "2026-10-01",
		CycleEnd:        "2026-10-31",
		AllocatedAmount: 0,
	}

	a, err := svc.CreateAllocation(context.Background(), uuid.New(), input)
	if err != nil {
		t.Fatalf("expected no error for zero allocation with zero income, got %v", err)
	}
	if a.AllocatedAmount != 0 {
		t.Errorf("expected 0, got %d", a.AllocatedAmount)
	}
}
