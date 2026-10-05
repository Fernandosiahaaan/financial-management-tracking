package transaction

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service defines business logic for transactions.
type Service struct {
	repo Repository
}

// NewService instantiates a new transaction Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateTransaction validates input and executes transaction creation.
func (s *Service) CreateTransaction(ctx context.Context, userID uuid.UUID, req CreateTransactionRequest) (*Transaction, error) {
	if req.AccountID == uuid.Nil {
		return nil, ErrMissingAccount
	}

	if !req.Type.Valid() {
		return nil, ErrInvalidType
	}

	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	req.TransactionDate = strings.TrimSpace(req.TransactionDate)
	if req.TransactionDate == "" {
		req.TransactionDate = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", req.TransactionDate); err != nil {
			return nil, ErrInvalidDate
		}
	}

	if req.Type == TypeTransfer {
		if req.DestinationAccountID == nil || *req.DestinationAccountID == uuid.Nil {
			return nil, ErrMissingDestAccount
		}
		if *req.DestinationAccountID == req.AccountID {
			return nil, ErrSameAccountTransfer
		}
	} else {
		if req.DestinationAccountID != nil && *req.DestinationAccountID != uuid.Nil {
			return nil, ErrTransferDestinationSet
		}
	}

	tx := Transaction{
		UserID:               userID,
		AccountID:            req.AccountID,
		DestinationAccountID: req.DestinationAccountID,
		CategoryID:           req.CategoryID,
		Type:                 req.Type,
		Amount:               req.Amount.Int64(),
		TransactionDate:      req.TransactionDate,
		Description:          strings.TrimSpace(req.Description),
	}

	return s.repo.Create(ctx, tx)
}

// GetTransaction retrieves a transaction by ID.
func (s *Service) GetTransaction(ctx context.Context, userID, id uuid.UUID) (*Transaction, error) {
	if id == uuid.Nil {
		return nil, ErrTransactionNotFound
	}
	return s.repo.GetByID(ctx, userID, id)
}

// ListTransactions returns filtered transactions for the user.
func (s *Service) ListTransactions(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Transaction, error) {
	if filter.StartDate != "" {
		if _, err := time.Parse("2006-01-02", filter.StartDate); err != nil {
			return nil, fmt.Errorf("start_date %w", ErrInvalidDate)
		}
	}
	if filter.EndDate != "" {
		if _, err := time.Parse("2006-01-02", filter.EndDate); err != nil {
			return nil, fmt.Errorf("end_date %w", ErrInvalidDate)
		}
	}

	return s.repo.List(ctx, userID, filter)
}

// UpdateTransaction validates and applies updates to an existing transaction.
func (s *Service) UpdateTransaction(ctx context.Context, userID, id uuid.UUID, req UpdateTransactionRequest) (*Transaction, error) {
	if id == uuid.Nil {
		return nil, ErrTransactionNotFound
	}
	if req.AccountID == uuid.Nil {
		return nil, ErrMissingAccount
	}
	if !req.Type.Valid() {
		return nil, ErrInvalidType
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	req.TransactionDate = strings.TrimSpace(req.TransactionDate)
	if req.TransactionDate == "" {
		req.TransactionDate = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", req.TransactionDate); err != nil {
			return nil, ErrInvalidDate
		}
	}

	if req.Type == TypeTransfer {
		if req.DestinationAccountID == nil || *req.DestinationAccountID == uuid.Nil {
			return nil, ErrMissingDestAccount
		}
		if *req.DestinationAccountID == req.AccountID {
			return nil, ErrSameAccountTransfer
		}
	} else {
		if req.DestinationAccountID != nil && *req.DestinationAccountID != uuid.Nil {
			return nil, ErrTransferDestinationSet
		}
	}

	newTx := Transaction{
		UserID:               userID,
		AccountID:            req.AccountID,
		DestinationAccountID: req.DestinationAccountID,
		CategoryID:           req.CategoryID,
		Type:                 req.Type,
		Amount:               req.Amount.Int64(),
		TransactionDate:      req.TransactionDate,
		Description:          strings.TrimSpace(req.Description),
	}

	return s.repo.Update(ctx, userID, id, newTx)
}

// DeleteTransaction soft-deletes a transaction and reverses its balance adjustments.
func (s *Service) DeleteTransaction(ctx context.Context, userID, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrTransactionNotFound
	}
	return s.repo.SoftDelete(ctx, userID, id)
}
