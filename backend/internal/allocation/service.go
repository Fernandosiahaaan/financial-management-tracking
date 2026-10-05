package allocation

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Service defines allocation business operations.
type Service interface {
	CreateAllocation(ctx context.Context, userID uuid.UUID, input CreateAllocationInput) (*Allocation, error)
	ListAllocations(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) (*AllocationSummary, error)
	GetAllocation(ctx context.Context, userID, id uuid.UUID) (*Allocation, error)
	UpdateAllocation(ctx context.Context, userID, id uuid.UUID, input UpdateAllocationInput) (*Allocation, error)
	DeleteAllocation(ctx context.Context, userID, id uuid.UUID) error
}

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// NewService creates a new allocation Service.
func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

func formatAllocation(a *Allocation) {
	if a != nil {
		a.FormattedAllocatedAmount = money.FormatIDR(a.AllocatedAmount)
	}
}

func (s *DefaultService) CreateAllocation(ctx context.Context, userID uuid.UUID, input CreateAllocationInput) (*Allocation, error) {
	if input.CategoryID == uuid.Nil {
		return nil, ErrInvalidCategoryID
	}

	if input.AllocatedAmount < 0 {
		return nil, ErrInvalidAllocatedAmount
	}

	cycleStart, cycleEnd, err := validateCycleDates(input.CycleStart, input.CycleEnd)
	if err != nil {
		return nil, err
	}

	// Verify category belongs to user
	belongs, err := s.repo.CategoryBelongsToUser(ctx, userID, input.CategoryID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, ErrCategoryNotFound
	}

	// Validate allocation doesn't exceed available income
	totalIncome, err := s.repo.GetCycleIncome(ctx, userID, cycleStart, cycleEnd)
	if err != nil {
		return nil, err
	}

	currentAllocated, err := s.repo.GetTotalAllocated(ctx, userID, cycleStart, cycleEnd, nil)
	if err != nil {
		return nil, err
	}

	if currentAllocated+input.AllocatedAmount.Int64() > totalIncome {
		return nil, ErrExceedsAvailableIncome
	}

	a := Allocation{
		UserID:          userID,
		CategoryID:      input.CategoryID,
		CycleStart:      cycleStart,
		CycleEnd:        cycleEnd,
		AllocatedAmount: input.AllocatedAmount.Int64(),
	}

	created, err := s.repo.Create(ctx, a)
	if err != nil {
		return nil, err
	}
	formatAllocation(created)
	return created, nil
}

func (s *DefaultService) ListAllocations(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) (*AllocationSummary, error) {
	cs, ce, err := validateCycleDates(cycleStart, cycleEnd)
	if err != nil {
		return nil, err
	}

	allocations, err := s.repo.ListByUserAndCycle(ctx, userID, cs, ce)
	if err != nil {
		return nil, err
	}
	for i := range allocations {
		formatAllocation(&allocations[i])
	}

	totalAllocated, err := s.repo.GetTotalAllocated(ctx, userID, cs, ce, nil)
	if err != nil {
		return nil, err
	}

	totalIncome, err := s.repo.GetCycleIncome(ctx, userID, cs, ce)
	if err != nil {
		return nil, err
	}

	remainingIncome := totalIncome - totalAllocated

	return &AllocationSummary{
		Allocations:              allocations,
		TotalAllocated:           totalAllocated,
		FormattedTotalAllocated:  money.FormatIDR(totalAllocated),
		TotalIncome:              totalIncome,
		FormattedTotalIncome:     money.FormatIDR(totalIncome),
		RemainingIncome:          remainingIncome,
		FormattedRemainingIncome: money.FormatIDR(remainingIncome),
	}, nil
}

func (s *DefaultService) GetAllocation(ctx context.Context, userID, id uuid.UUID) (*Allocation, error) {
	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	formatAllocation(a)
	return a, nil
}

func (s *DefaultService) UpdateAllocation(ctx context.Context, userID, id uuid.UUID, input UpdateAllocationInput) (*Allocation, error) {
	if input.AllocatedAmount < 0 {
		return nil, ErrInvalidAllocatedAmount
	}

	// Get existing allocation to find cycle dates
	existing, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	// Validate new total won't exceed income (exclude current allocation from sum)
	totalIncome, err := s.repo.GetCycleIncome(ctx, userID, existing.CycleStart, existing.CycleEnd)
	if err != nil {
		return nil, err
	}

	otherAllocated, err := s.repo.GetTotalAllocated(ctx, userID, existing.CycleStart, existing.CycleEnd, &id)
	if err != nil {
		return nil, err
	}

	if otherAllocated+input.AllocatedAmount.Int64() > totalIncome {
		return nil, ErrExceedsAvailableIncome
	}

	updated, err := s.repo.Update(ctx, userID, id, input.AllocatedAmount.Int64())
	if err != nil {
		return nil, err
	}
	formatAllocation(updated)
	return updated, nil
}

func (s *DefaultService) DeleteAllocation(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, userID, id)
}

// validateCycleDates parses and validates cycle date strings.
func validateCycleDates(startStr, endStr string) (string, string, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return "", "", ErrInvalidCycleDates
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return "", "", ErrInvalidCycleDates
	}

	if !start.Before(end) {
		return "", "", ErrInvalidCycleDates
	}

	return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
}
