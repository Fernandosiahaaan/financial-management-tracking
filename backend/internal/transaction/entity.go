package transaction

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Type represents the financial classification of a transaction.
type Type string

const (
	TypeIncome   Type = "INCOME"
	TypeExpense  Type = "EXPENSE"
	TypeTransfer Type = "TRANSFER"
)

// Valid returns true if the TransactionType is supported.
func (t Type) Valid() bool {
	switch t {
	case TypeIncome, TypeExpense, TypeTransfer:
		return true
	default:
		return false
	}
}

// Transaction represents a financial event affecting one or two accounts.
type Transaction struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	AccountID            uuid.UUID  `json:"account_id"`
	DestinationAccountID *uuid.UUID `json:"destination_account_id,omitempty"`
	CategoryID           *uuid.UUID `json:"category_id,omitempty"`
	Type                 Type       `json:"type"`
	Amount               int64      `json:"amount"` // In cents
	FormattedAmount      string     `json:"formatted_amount,omitempty"`
	TransactionDate      string     `json:"transaction_date"` // YYYY-MM-DD
	Description          string     `json:"description,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`

	// Joined fields for display
	AccountName            string `json:"account_name,omitempty"`
	DestinationAccountName string `json:"destination_account_name,omitempty"`
	CategoryName           string `json:"category_name,omitempty"`
	CategoryColor          string `json:"category_color,omitempty"`
	CategoryIcon           string `json:"category_icon,omitempty"`
}

// CreateTransactionRequest defines payload for creating a transaction.
type CreateTransactionRequest struct {
	AccountID            uuid.UUID   `json:"account_id"`
	DestinationAccountID *uuid.UUID  `json:"destination_account_id,omitempty"`
	CategoryID           *uuid.UUID  `json:"category_id,omitempty"`
	Type                 Type        `json:"type"`
	Amount               money.Cents `json:"amount"`
	TransactionDate      string      `json:"transaction_date"` // YYYY-MM-DD
	Description          string      `json:"description"`
}

// UpdateTransactionRequest defines payload for modifying an existing transaction.
type UpdateTransactionRequest struct {
	AccountID            uuid.UUID   `json:"account_id"`
	DestinationAccountID *uuid.UUID  `json:"destination_account_id,omitempty"`
	CategoryID           *uuid.UUID  `json:"category_id,omitempty"`
	Type                 Type        `json:"type"`
	Amount               money.Cents `json:"amount"`
	TransactionDate      string      `json:"transaction_date"`
	Description          string      `json:"description"`
}

// ListFilter specifies query parameters for filtering transactions.
type ListFilter struct {
	StartDate string     `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate   string     `json:"end_date,omitempty"`   // YYYY-MM-DD
	Type      *Type      `json:"type,omitempty"`
	AccountID *uuid.UUID `json:"account_id,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

var (
	ErrInvalidType             = errors.New("invalid transaction type: must be INCOME, EXPENSE, or TRANSFER")
	ErrInvalidAmount           = errors.New("amount must be greater than zero")
	ErrInvalidDate             = errors.New("invalid transaction date: format must be YYYY-MM-DD")
	ErrMissingAccount          = errors.New("account_id is required")
	ErrMissingDestAccount      = errors.New("destination_account_id is required for transfer transactions")
	ErrSameAccountTransfer     = errors.New("source and destination accounts must be different for transfers")
	ErrTransferDestinationSet  = errors.New("destination_account_id should not be set for income or expense transactions")
	ErrTransactionNotFound     = errors.New("transaction not found")
	ErrInsufficientBalance     = errors.New("insufficient balance in source account")
	ErrAccountNotFound         = errors.New("specified account does not exist or does not belong to user")
	ErrCategoryNotFound        = errors.New("specified category does not exist or does not belong to user")
)
