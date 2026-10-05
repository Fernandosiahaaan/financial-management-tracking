package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockRepository struct {
	createFn     func(ctx context.Context, tx Transaction) (*Transaction, error)
	getByIDFn    func(ctx context.Context, userID, id uuid.UUID) (*Transaction, error)
	listFn       func(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Transaction, error)
	updateFn     func(ctx context.Context, userID, id uuid.UUID, newTx Transaction) (*Transaction, error)
	softDeleteFn func(ctx context.Context, userID, id uuid.UUID) error
}

func (m *mockRepository) Create(ctx context.Context, tx Transaction) (*Transaction, error) {
	if m.createFn != nil {
		return m.createFn(ctx, tx)
	}
	return &tx, nil
}

func (m *mockRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Transaction, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return &Transaction{ID: id, UserID: userID}, nil
}

func (m *mockRepository) List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Transaction, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, filter)
	}
	return []Transaction{}, nil
}

func (m *mockRepository) Update(ctx context.Context, userID, id uuid.UUID, newTx Transaction) (*Transaction, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, userID, id, newTx)
	}
	return &newTx, nil
}

func (m *mockRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	if m.softDeleteFn != nil {
		return m.softDeleteFn(ctx, userID, id)
	}
	return nil
}

func TestTransactionService_Validation(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	ctx := context.Background()
	userID := uuid.New()
	accID := uuid.New()
	destID := uuid.New()

	tests := []struct {
		name        string
		req         CreateTransactionRequest
		expectedErr error
	}{
		{
			name: "missing account ID",
			req: CreateTransactionRequest{
				AccountID: uuid.Nil,
				Type:      TypeExpense,
				Amount:    50000,
			},
			expectedErr: ErrMissingAccount,
		},
		{
			name: "invalid transaction type",
			req: CreateTransactionRequest{
				AccountID: accID,
				Type:      "INVALID",
				Amount:    50000,
			},
			expectedErr: ErrInvalidType,
		},
		{
			name: "zero amount",
			req: CreateTransactionRequest{
				AccountID: accID,
				Type:      TypeExpense,
				Amount:    0,
			},
			expectedErr: ErrInvalidAmount,
		},
		{
			name: "negative amount",
			req: CreateTransactionRequest{
				AccountID: accID,
				Type:      TypeExpense,
				Amount:    -10000,
			},
			expectedErr: ErrInvalidAmount,
		},
		{
			name: "invalid date format",
			req: CreateTransactionRequest{
				AccountID:       accID,
				Type:            TypeIncome,
				Amount:          100000,
				TransactionDate: "05-10-2026",
			},
			expectedErr: ErrInvalidDate,
		},
		{
			name: "transfer missing destination account",
			req: CreateTransactionRequest{
				AccountID:       accID,
				Type:            TypeTransfer,
				Amount:          20000,
				TransactionDate: "2026-10-05",
			},
			expectedErr: ErrMissingDestAccount,
		},
		{
			name: "transfer same account",
			req: CreateTransactionRequest{
				AccountID:            accID,
				DestinationAccountID: &accID,
				Type:                 TypeTransfer,
				Amount:               20000,
				TransactionDate:      "2026-10-05",
			},
			expectedErr: ErrSameAccountTransfer,
		},
		{
			name: "expense with destination account set",
			req: CreateTransactionRequest{
				AccountID:            accID,
				DestinationAccountID: &destID,
				Type:                 TypeExpense,
				Amount:               20000,
				TransactionDate:      "2026-10-05",
			},
			expectedErr: ErrTransferDestinationSet,
		},
		{
			name: "valid income transaction",
			req: CreateTransactionRequest{
				AccountID:       accID,
				Type:            TypeIncome,
				Amount:          5000000,
				TransactionDate: "2026-10-01",
				Description:     "Salary October",
			},
			expectedErr: nil,
		},
		{
			name: "valid transfer transaction",
			req: CreateTransactionRequest{
				AccountID:            accID,
				DestinationAccountID: &destID,
				Type:                 TypeTransfer,
				Amount:               150000,
				TransactionDate:      "2026-10-05",
				Description:          "Wallet topup",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := svc.CreateTransaction(ctx, userID, tt.req)
			if tt.expectedErr != nil {
				if err == nil || err != tt.expectedErr {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if created.Amount != tt.req.Amount.Int64() {
					t.Errorf("expected amount %d, got %d", tt.req.Amount.Int64(), created.Amount)
				}
			}
		})
	}
}

func TestTransactionService_UpdateAndDelete(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	ctx := context.Background()
	userID := uuid.New()
	txID := uuid.New()
	accID := uuid.New()

	// Update with invalid amount
	_, err := svc.UpdateTransaction(ctx, userID, txID, UpdateTransactionRequest{
		AccountID: accID,
		Type:      TypeExpense,
		Amount:    -500,
	})
	if err != ErrInvalidAmount {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}

	// Update with valid data
	updated, err := svc.UpdateTransaction(ctx, userID, txID, UpdateTransactionRequest{
		AccountID:       accID,
		Type:            TypeExpense,
		Amount:          25000,
		TransactionDate: "2026-10-05",
		Description:     "Lunch",
	})
	if err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}
	if updated.Amount != 25000 {
		t.Errorf("expected amount 25000, got %d", updated.Amount)
	}

	// Delete
	err = svc.DeleteTransaction(ctx, userID, txID)
	if err != nil {
		t.Fatalf("unexpected error on delete: %v", err)
	}
}

func TestTransactionService_ListFilterValidation(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	ctx := context.Background()
	userID := uuid.New()

	// Invalid start_date
	_, err := svc.ListTransactions(ctx, userID, ListFilter{StartDate: "not-a-date"})
	if err == nil {
		t.Fatal("expected error on invalid start_date, got nil")
	}

	// Invalid end_date
	_, err = svc.ListTransactions(ctx, userID, ListFilter{EndDate: "invalid-date"})
	if err == nil {
		t.Fatal("expected error on invalid end_date, got nil")
	}

	// Valid dates
	list, err := svc.ListTransactions(ctx, userID, ListFilter{
		StartDate: "2026-09-01",
		EndDate:   "2026-10-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if list == nil {
		t.Fatal("expected empty slice, got nil")
	}
}
