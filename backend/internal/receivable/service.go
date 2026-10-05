package receivable

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service defines business use-cases for receivables.
type Service struct {
	repo Repository
}

// NewService creates a new receivable Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create validates and creates a new receivable.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateReceivableRequest) (*Receivable, error) {
	counterparty := strings.TrimSpace(req.Counterparty)
	if counterparty == "" {
		return nil, ErrInvalidCounterparty
	}

	if req.Principal <= 0 {
		return nil, ErrInvalidPrincipal
	}

	var dueDateVal *string
	if req.DueDate != nil && strings.TrimSpace(*req.DueDate) != "" {
		trimmed := strings.TrimSpace(*req.DueDate)
		if _, err := time.Parse("2006-01-02", trimmed); err != nil {
			return nil, ErrInvalidDueDate
		}
		dueDateVal = &trimmed
	}

	rec := Receivable{
		UserID:          userID,
		SourceAccountID: req.SourceAccountID,
		Counterparty:    counterparty,
		Principal:       req.Principal.Int64(),
		Status:          StatusActive,
		DueDate:         dueDateVal,
		Notes:           strings.TrimSpace(req.Notes),
	}

	return s.repo.Create(ctx, rec)
}

// GetByID retrieves a single receivable with payment history.
func (s *Service) GetByID(ctx context.Context, userID, id uuid.UUID) (*Receivable, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// List returns user receivables matching optional filter.
func (s *Service) List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Receivable, error) {
	if filter.Status != nil && !filter.Status.Valid() {
		return nil, ErrInvalidStatus
	}
	return s.repo.List(ctx, userID, filter)
}

// RecordPayment validates and records a partial or full repayment.
func (s *Service) RecordPayment(ctx context.Context, userID, receivableID uuid.UUID, req RecordPaymentRequest) (*ReceivablePayment, *Receivable, error) {
	if req.Amount <= 0 {
		return nil, nil, ErrInvalidPaymentAmount
	}

	paymentDate := strings.TrimSpace(req.PaymentDate)
	if paymentDate == "" {
		return nil, nil, ErrInvalidPaymentDate
	}
	if _, err := time.Parse("2006-01-02", paymentDate); err != nil {
		return nil, nil, ErrInvalidPaymentDate
	}

	if req.TargetAccountID == uuid.Nil {
		return nil, nil, ErrMissingTargetAccount
	}

	payment := ReceivablePayment{
		ReceivableID:    receivableID,
		UserID:          userID,
		TargetAccountID: req.TargetAccountID,
		Amount:          req.Amount.Int64(),
		PaymentDate:     paymentDate,
		Notes:           strings.TrimSpace(req.Notes),
	}

	return s.repo.RecordPayment(ctx, payment)
}

// UpdateStatus modifies the receivable status (e.g. marking as WRITTEN_OFF or OVERDUE).
func (s *Service) UpdateStatus(ctx context.Context, userID, id uuid.UUID, req UpdateStatusRequest) (*Receivable, error) {
	if !req.Status.Valid() {
		return nil, ErrInvalidStatus
	}

	var notesVal *string
	if req.Notes != nil {
		trimmed := strings.TrimSpace(*req.Notes)
		notesVal = &trimmed
	}

	return s.repo.UpdateStatus(ctx, userID, id, req.Status, notesVal)
}

// SoftDelete removes a receivable.
func (s *Service) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, userID, id)
}

// GetTotalActiveReceivables computes total outstanding unpaid principal across active receivables.
func (s *Service) GetTotalActiveReceivables(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.GetTotalActiveReceivables(ctx, userID)
}
