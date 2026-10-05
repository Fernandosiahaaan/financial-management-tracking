package investment

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

type mockRepository struct {
	investments []Investment
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
	totalValue  int64
}

func (m *mockRepository) Create(_ context.Context, inv Investment) (*Investment, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	inv.ID = uuid.New()
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = time.Now()
	inv.UnrealizedGain = inv.CurrentValue - inv.Capital
	if inv.Capital > 0 {
		inv.UnrealizedGainPercentage = (float64(inv.UnrealizedGain) / float64(inv.Capital)) * 100
	}
	m.investments = append(m.investments, inv)
	return &inv, nil
}

func (m *mockRepository) GetByID(_ context.Context, userID, id uuid.UUID) (*Investment, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, inv := range m.investments {
		if inv.ID == id && inv.UserID == userID {
			return &inv, nil
		}
	}
	return nil, ErrInvestmentNotFound
}

func (m *mockRepository) List(_ context.Context, userID uuid.UUID, filter ListFilter) ([]Investment, error) {
	var result []Investment
	for _, inv := range m.investments {
		if inv.UserID == userID {
			if filter.Type != nil && inv.Type != *filter.Type {
				continue
			}
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *mockRepository) UpdateValuation(_ context.Context, userID, id uuid.UUID, currentValue int64, notes *string) (*Investment, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	for i, inv := range m.investments {
		if inv.ID == id && inv.UserID == userID {
			m.investments[i].CurrentValue = currentValue
			m.investments[i].UnrealizedGain = currentValue - m.investments[i].Capital
			if m.investments[i].Capital > 0 {
				m.investments[i].UnrealizedGainPercentage = (float64(m.investments[i].UnrealizedGain) / float64(m.investments[i].Capital)) * 100
			}
			if notes != nil {
				m.investments[i].Notes = *notes
			}
			return &m.investments[i], nil
		}
	}
	return nil, ErrInvestmentNotFound
}

func (m *mockRepository) SoftDelete(_ context.Context, userID, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, inv := range m.investments {
		if inv.ID == id && inv.UserID == userID {
			now := time.Now()
			m.investments[i].DeletedAt = &now
			return nil
		}
	}
	return ErrInvestmentNotFound
}

func (m *mockRepository) GetTotalInvestmentsValue(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.totalValue, nil
}

func TestInvestmentService_Create(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	userID := uuid.New()

	// Empty name
	_, err := svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:    "   ",
		Type:    TypeStock,
		Capital: 10000000,
	})
	if err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}

	// Invalid type
	_, err = svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:    "BBCA",
		Type:    Type("REAL_ESTATE"),
		Capital: 10000000,
	})
	if err != ErrInvalidType {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}

	// Negative capital
	_, err = svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:    "BBCA",
		Type:    TypeStock,
		Capital: -500,
	})
	if err != ErrInvalidCapital {
		t.Fatalf("expected ErrInvalidCapital, got %v", err)
	}

	// Successful creation with default valuation = capital
	inv, err := svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:    "Bank Central Asia",
		Type:    TypeStock,
		Capital: 15000000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Name != "Bank Central Asia" || inv.Capital != 15000000 || inv.CurrentValue != 15000000 {
		t.Errorf("unexpected investment fields: %+v", inv)
	}
}

func TestInvestmentService_ListAndPortfolioSummary(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	userID := uuid.New()

	val1 := money.Cents(12000000)
	_, _ = svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:         "BBCA",
		Type:         TypeStock,
		Capital:      10000000,
		CurrentValue: &val1, // +2,000,000 (+20%)
	})

	val2 := money.Cents(4500000)
	_, _ = svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:         "Antam Gold 5g",
		Type:         TypeGold,
		Capital:      5000000,
		CurrentValue: &val2, // -500,000 (-10%)
	})

	summary, err := svc.List(context.Background(), userID, ListFilter{})
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if summary.InvestmentsCount != 2 {
		t.Errorf("expected 2 investments, got %d", summary.InvestmentsCount)
	}
	if summary.TotalCapital != 15000000 {
		t.Errorf("expected 15000000 total capital, got %d", summary.TotalCapital)
	}
	if summary.TotalCurrentValue != 16500000 {
		t.Errorf("expected 16500000 total current value, got %d", summary.TotalCurrentValue)
	}
	if summary.TotalGainLoss != 1500000 {
		t.Errorf("expected 1500000 total gain/loss, got %d", summary.TotalGainLoss)
	}
	expectedPct := 10.0 // +1.5M / 15M = 10%
	if summary.TotalGainLossPercentage != expectedPct {
		t.Errorf("expected 10.0%% gain percentage, got %f", summary.TotalGainLossPercentage)
	}
}

func TestInvestmentService_UpdateValuation(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	userID := uuid.New()

	inv, _ := svc.Create(context.Background(), userID, CreateInvestmentRequest{
		Name:    "Bitcoin",
		Type:    TypeCrypto,
		Capital: 20000000,
	})

	// Negative valuation
	_, err := svc.UpdateValuation(context.Background(), userID, inv.ID, UpdateValuationRequest{
		CurrentValue: -1,
	})
	if err != ErrInvalidCurrentValue {
		t.Fatalf("expected ErrInvalidCurrentValue, got %v", err)
	}

	// Successful valuation update
	notes := "Market rally"
	updated, err := svc.UpdateValuation(context.Background(), userID, inv.ID, UpdateValuationRequest{
		CurrentValue: 25000000,
		Notes:        &notes,
	})
	if err != nil {
		t.Fatalf("unexpected update valuation error: %v", err)
	}
	if updated.CurrentValue != 25000000 || updated.UnrealizedGain != 5000000 || updated.UnrealizedGainPercentage != 25.0 {
		t.Errorf("unexpected updated investment: %+v", updated)
	}
}
