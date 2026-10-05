package budget

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Service defines budget business operations.
type Service interface {
	CreateBudget(ctx context.Context, userID uuid.UUID, input CreateBudgetInput) (*Budget, error)
	ListBudgets(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]Budget, error)
	GetBudget(ctx context.Context, userID, id uuid.UUID) (*Budget, error)
	UpdateBudget(ctx context.Context, userID, id uuid.UUID, input UpdateBudgetInput) (*Budget, error)
	DeleteBudget(ctx context.Context, userID, id uuid.UUID) error
}

// DefaultService implements Service.
type DefaultService struct {
	repo Repository
}

// NewService creates a new budget Service.
func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

func formatBudget(b *Budget) {
	b.FormattedPlannedAmount = money.FormatIDR(b.PlannedAmount)
	b.FormattedActualAmount = money.FormatIDR(b.ActualAmount)
	b.FormattedVariance = money.FormatIDR(b.Variance)
}

func (s *DefaultService) CreateBudget(ctx context.Context, userID uuid.UUID, input CreateBudgetInput) (*Budget, error) {
	if input.CategoryID == uuid.Nil {
		return nil, ErrInvalidCategoryID
	}

	if input.PlannedAmount < 0 {
		return nil, ErrInvalidPlannedAmount
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

	b := Budget{
		UserID:        userID,
		CategoryID:    input.CategoryID,
		CycleStart:    cycleStart,
		CycleEnd:      cycleEnd,
		PlannedAmount: input.PlannedAmount.Int64(),
	}

	created, err := s.repo.Create(ctx, b)
	if err != nil {
		return nil, err
	}

	// Enrich with actual spending
	actual, err := s.repo.GetActualSpending(ctx, userID, created.CategoryID, created.CycleStart, created.CycleEnd)
	if err != nil {
		return nil, err
	}
	created.ActualAmount = actual
	created.Variance = created.PlannedAmount - actual
	formatBudget(created)

	return created, nil
}

func (s *DefaultService) ListBudgets(ctx context.Context, userID uuid.UUID, cycleStart, cycleEnd string) ([]Budget, error) {
	cs, ce, err := validateCycleDates(cycleStart, cycleEnd)
	if err != nil {
		return nil, err
	}

	budgets, err := s.repo.ListByUserAndCycle(ctx, userID, cs, ce)
	if err != nil {
		return nil, err
	}

	// Enrich each budget with actual spending and variance
	for i := range budgets {
		actual, err := s.repo.GetActualSpending(ctx, userID, budgets[i].CategoryID, cs, ce)
		if err != nil {
			return nil, err
		}
		budgets[i].ActualAmount = actual
		budgets[i].Variance = budgets[i].PlannedAmount - actual
		formatBudget(&budgets[i])
	}

	return budgets, nil
}

func (s *DefaultService) GetBudget(ctx context.Context, userID, id uuid.UUID) (*Budget, error) {
	b, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	actual, err := s.repo.GetActualSpending(ctx, userID, b.CategoryID, b.CycleStart, b.CycleEnd)
	if err != nil {
		return nil, err
	}
	b.ActualAmount = actual
	b.Variance = b.PlannedAmount - actual
	formatBudget(b)

	return b, nil
}

func (s *DefaultService) UpdateBudget(ctx context.Context, userID, id uuid.UUID, input UpdateBudgetInput) (*Budget, error) {
	if input.PlannedAmount < 0 {
		return nil, ErrInvalidPlannedAmount
	}

	updated, err := s.repo.Update(ctx, userID, id, input.PlannedAmount.Int64())
	if err != nil {
		return nil, err
	}

	actual, err := s.repo.GetActualSpending(ctx, userID, updated.CategoryID, updated.CycleStart, updated.CycleEnd)
	if err != nil {
		return nil, err
	}
	updated.ActualAmount = actual
	updated.Variance = updated.PlannedAmount - actual
	formatBudget(updated)

	return updated, nil
}

func (s *DefaultService) DeleteBudget(ctx context.Context, userID, id uuid.UUID) error {
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
