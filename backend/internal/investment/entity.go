package investment

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Type represents the asset classification of an investment.
type Type string

const (
	TypeStock      Type = "STOCK"
	TypeMutualFund Type = "MUTUAL_FUND"
	TypeGold       Type = "GOLD"
	TypeCrypto     Type = "CRYPTO"
	TypeOther      Type = "OTHER"
)

// Valid checks if the investment type is supported.
func (t Type) Valid() bool {
	switch t {
	case TypeStock, TypeMutualFund, TypeGold, TypeCrypto, TypeOther:
		return true
	default:
		return false
	}
}

// Investment represents an investment holding.
type Investment struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   uuid.UUID  `json:"user_id"`
	SourceAccountID          *uuid.UUID `json:"source_account_id,omitempty"`
	SourceAccountName        string     `json:"source_account_name,omitempty"`
	Type                     Type       `json:"type"`
	Name                     string     `json:"name"`
	Capital                  int64      `json:"capital"`       // Capital invested in cents
	FormattedCapital         string     `json:"formatted_capital,omitempty"`
	CurrentValue             int64      `json:"current_value"` // Latest valuation in cents
	FormattedCurrentValue    string     `json:"formatted_current_value,omitempty"`
	UnrealizedGain           int64      `json:"unrealized_gain"`
	FormattedUnrealizedGain  string     `json:"formatted_unrealized_gain,omitempty"`
	UnrealizedGainPercentage float64    `json:"unrealized_gain_percentage"`
	Notes                    string     `json:"notes"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	DeletedAt                *time.Time `json:"deleted_at,omitempty"`
}

// PortfolioSummary aggregates portfolio metrics across holdings.
type PortfolioSummary struct {
	TotalCapital                int64        `json:"total_capital"`
	FormattedTotalCapital       string       `json:"formatted_total_capital,omitempty"`
	TotalCurrentValue           int64        `json:"total_current_value"`
	FormattedTotalCurrentValue  string       `json:"formatted_total_current_value,omitempty"`
	TotalGainLoss               int64        `json:"total_gain_loss"`
	FormattedTotalGainLoss      string       `json:"formatted_total_gain_loss,omitempty"`
	TotalGainLossPercentage     float64      `json:"total_gain_loss_percentage"`
	InvestmentsCount            int          `json:"investments_count"`
	Investments                 []Investment `json:"investments"`
}

// CreateInvestmentRequest represents payload to create an investment.
type CreateInvestmentRequest struct {
	Type            Type         `json:"type"`
	Name            string       `json:"name"`
	Capital         money.Cents  `json:"capital"`
	CurrentValue    *money.Cents `json:"current_value,omitempty"` // Defaults to Capital if nil
	SourceAccountID *uuid.UUID   `json:"source_account_id,omitempty"`
	Notes           string       `json:"notes,omitempty"`
}

// UpdateValuationRequest represents payload to update an investment's valuation.
type UpdateValuationRequest struct {
	CurrentValue money.Cents `json:"current_value"`
	Notes        *string     `json:"notes,omitempty"`
}

// ListFilter specifies filters for investments.
type ListFilter struct {
	Type *Type `json:"type,omitempty"`
}

var (
	ErrInvalidType         = errors.New("invalid investment type: must be STOCK, MUTUAL_FUND, GOLD, CRYPTO, or OTHER")
	ErrInvalidName         = errors.New("investment name is required")
	ErrInvalidCapital      = errors.New("capital must be non-negative")
	ErrInvalidCurrentValue = errors.New("current value must be non-negative")
	ErrInvestmentNotFound  = errors.New("investment not found")
	ErrAccountNotFound     = errors.New("source account not found or does not belong to user")
	ErrInsufficientBalance = errors.New("insufficient balance in source account")
)
