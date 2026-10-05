package account

import (
	"time"

	"github.com/google/uuid"
)

// Type represents the financial classification of an account.
type Type string

const (
	TypeBank       Type = "BANK"
	TypeEWallet    Type = "E_WALLET"
	TypeCash       Type = "CASH"
	TypeInvestment Type = "INVESTMENT"
	TypeOther      Type = "OTHER"
)

// Valid returns true if the AccountType is supported.
func (t Type) Valid() bool {
	switch t {
	case TypeBank, TypeEWallet, TypeCash, TypeInvestment, TypeOther:
		return true
	default:
		return false
	}
}

// Status represents the operational state of an account.
type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
)

// Valid returns true if the Status is supported.
func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusArchived:
		return true
	default:
		return false
	}
}

// Account represents a financial store of value scoped to a user.
type Account struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	Name           string     `json:"name"`
	Type           Type       `json:"type"`
	OpeningBalance   int64      `json:"opening_balance"`
	CurrentBalance   int64      `json:"current_balance"`
	FormattedBalance string     `json:"formatted_balance,omitempty"`
	Status           Status     `json:"status"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
