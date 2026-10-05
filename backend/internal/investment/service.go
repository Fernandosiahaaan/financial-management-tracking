package investment

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Service coordinates business logic for investments.
type Service struct {
	repo Repository
}

// NewService creates a new investment Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create validates and creates an investment holding.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateInvestmentRequest) (*Investment, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidName
	}

	if !req.Type.Valid() {
		return nil, ErrInvalidType
	}

	if req.Capital < 0 {
		return nil, ErrInvalidCapital
	}

	currentVal := req.Capital.Int64()
	if req.CurrentValue != nil {
		if *req.CurrentValue < 0 {
			return nil, ErrInvalidCurrentValue
		}
		currentVal = req.CurrentValue.Int64()
	}

	inv := Investment{
		UserID:          userID,
		SourceAccountID: req.SourceAccountID,
		Type:            req.Type,
		Name:            name,
		Capital:         req.Capital.Int64(),
		CurrentValue:    currentVal,
		Notes:           strings.TrimSpace(req.Notes),
	}

	return s.repo.Create(ctx, inv)
}

// GetByID returns a single investment.
func (s *Service) GetByID(ctx context.Context, userID, id uuid.UUID) (*Investment, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// List returns user investments and computes portfolio summary.
func (s *Service) List(ctx context.Context, userID uuid.UUID, filter ListFilter) (*PortfolioSummary, error) {
	if filter.Type != nil && !filter.Type.Valid() {
		return nil, ErrInvalidType
	}

	items, err := s.repo.List(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	summary := &PortfolioSummary{
		InvestmentsCount: len(items),
		Investments:      items,
	}

	for _, item := range items {
		summary.TotalCapital += item.Capital
		summary.TotalCurrentValue += item.CurrentValue
	}

	summary.TotalGainLoss = summary.TotalCurrentValue - summary.TotalCapital
	if summary.TotalCapital > 0 {
		summary.TotalGainLossPercentage = (float64(summary.TotalGainLoss) / float64(summary.TotalCapital)) * 100
	}
	summary.FormattedTotalCapital = money.FormatIDR(summary.TotalCapital)
	summary.FormattedTotalCurrentValue = money.FormatIDR(summary.TotalCurrentValue)
	summary.FormattedTotalGainLoss = money.FormatIDR(summary.TotalGainLoss)

	return summary, nil
}

// UpdateValuation updates the market valuation of an investment.
func (s *Service) UpdateValuation(ctx context.Context, userID, id uuid.UUID, req UpdateValuationRequest) (*Investment, error) {
	if req.CurrentValue < 0 {
		return nil, ErrInvalidCurrentValue
	}

	var notesVal *string
	if req.Notes != nil {
		trimmed := strings.TrimSpace(*req.Notes)
		notesVal = &trimmed
	}

	return s.repo.UpdateValuation(ctx, userID, id, req.CurrentValue.Int64(), notesVal)
}

// SoftDelete marks an investment as deleted.
func (s *Service) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, userID, id)
}

// GetTotalInvestmentsValue returns total valuation across holdings.
func (s *Service) GetTotalInvestmentsValue(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.GetTotalInvestmentsValue(ctx, userID)
}
