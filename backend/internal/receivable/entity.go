package receivable

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Status represents the lifecycle status of a receivable.
type Status string

const (
	StatusActive        Status = "ACTIVE"
	StatusPartiallyPaid Status = "PARTIALLY_PAID"
	StatusPaid          Status = "PAID"
	StatusOverdue       Status = "OVERDUE"
	StatusWrittenOff    Status = "WRITTEN_OFF"
)

// Valid returns true if the status is a supported receivable status.
func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusPartiallyPaid, StatusPaid, StatusOverdue, StatusWrittenOff:
		return true
	default:
		return false
	}
}

// Receivable represents money lent to a counterparty.
type Receivable struct {
	ID                       uuid.UUID           `json:"id"`
	UserID                   uuid.UUID           `json:"user_id"`
	SourceAccountID          *uuid.UUID          `json:"source_account_id,omitempty"`
	SourceAccountName        string              `json:"source_account_name,omitempty"`
	Counterparty             string              `json:"counterparty"`
	Principal                int64               `json:"principal"` // In cents
	FormattedPrincipal       string              `json:"formatted_principal,omitempty"`
	TotalPaid                int64               `json:"total_paid"` // Aggregated from payments
	FormattedTotalPaid       string              `json:"formatted_total_paid,omitempty"`
	RemainingAmount          int64               `json:"remaining_amount"` // Principal - TotalPaid
	FormattedRemainingAmount string              `json:"formatted_remaining_amount,omitempty"`
	Status                   Status              `json:"status"`
	DueDate                  *string             `json:"due_date,omitempty"` // YYYY-MM-DD
	Notes                    string              `json:"notes"`
	CreatedAt                time.Time           `json:"created_at"`
	UpdatedAt                time.Time           `json:"updated_at"`
	DeletedAt                *time.Time          `json:"deleted_at,omitempty"`
	Payments                 []ReceivablePayment `json:"payments,omitempty"`
}

// ReceivablePayment represents a partial or full repayment for a receivable.
type ReceivablePayment struct {
	ID                uuid.UUID `json:"id"`
	ReceivableID      uuid.UUID `json:"receivable_id"`
	UserID            uuid.UUID `json:"user_id"`
	TargetAccountID   uuid.UUID `json:"target_account_id"`
	TargetAccountName string    `json:"target_account_name,omitempty"`
	Amount            int64     `json:"amount"` // Strictly positive in cents
	FormattedAmount   string    `json:"formatted_amount,omitempty"`
	PaymentDate       string    `json:"payment_date"` // YYYY-MM-DD
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateReceivableRequest represents the payload to create a new receivable.
type CreateReceivableRequest struct {
	Counterparty    string      `json:"counterparty"`
	Principal       money.Cents `json:"principal"`
	SourceAccountID *uuid.UUID  `json:"source_account_id,omitempty"`
	DueDate         *string     `json:"due_date,omitempty"`
	Notes           string      `json:"notes,omitempty"`
}

// RecordPaymentRequest represents the payload to record a repayment.
type RecordPaymentRequest struct {
	Amount          money.Cents `json:"amount"`
	TargetAccountID uuid.UUID   `json:"target_account_id"`
	PaymentDate     string      `json:"payment_date"` // YYYY-MM-DD
	Notes           string      `json:"notes,omitempty"`
}

// UpdateStatusRequest represents the payload to update a receivable's status (e.g. write off).
type UpdateStatusRequest struct {
	Status Status  `json:"status"`
	Notes  *string `json:"notes,omitempty"`
}

// ListFilter allows filtering receivables.
type ListFilter struct {
	Status *Status `json:"status,omitempty"`
}

var (
	ErrInvalidCounterparty  = errors.New("counterparty name is required")
	ErrInvalidPrincipal     = errors.New("principal must be greater than zero")
	ErrInvalidPaymentAmount = errors.New("payment amount must be greater than zero")
	ErrInvalidPaymentDate   = errors.New("payment date must be in YYYY-MM-DD format")
	ErrInvalidDueDate       = errors.New("due date must be in YYYY-MM-DD format")
	ErrOverpayment          = errors.New("payment amount exceeds remaining outstanding principal")
	ErrAlreadySettled       = errors.New("receivable is already settled or written off")
	ErrReceivableNotFound   = errors.New("receivable not found")
	ErrInvalidStatus        = errors.New("invalid receivable status")
	ErrAccountNotFound      = errors.New("specified account not found or does not belong to user")
	ErrInsufficientBalance  = errors.New("insufficient balance in source account")
	ErrMissingTargetAccount = errors.New("target account id is required for recording repayment")
)
