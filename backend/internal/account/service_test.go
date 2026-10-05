package account

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockAccountRepo struct {
	accounts map[uuid.UUID]Account
}

func newMockAccountRepo() *mockAccountRepo {
	return &mockAccountRepo{accounts: make(map[uuid.UUID]Account)}
}

func (m *mockAccountRepo) Create(ctx context.Context, acc Account) (*Account, error) {
	for _, a := range m.accounts {
		if a.UserID == acc.UserID && strings.EqualFold(a.Name, acc.Name) && a.DeletedAt == nil {
			return nil, ErrAccountAlreadyExists
		}
	}
	acc.ID = uuid.New()
	acc.CreatedAt = time.Now()
	acc.UpdatedAt = time.Now()
	m.accounts[acc.ID] = acc
	return &acc, nil
}

func (m *mockAccountRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]Account, error) {
	var list []Account
	for _, a := range m.accounts {
		if a.UserID == userID && a.DeletedAt == nil {
			list = append(list, a)
		}
	}
	return list, nil
}

func (m *mockAccountRepo) GetByID(ctx context.Context, userID, id uuid.UUID) (*Account, error) {
	a, ok := m.accounts[id]
	if !ok || a.UserID != userID || a.DeletedAt != nil {
		return nil, ErrAccountNotFound
	}
	return &a, nil
}

func (m *mockAccountRepo) Update(ctx context.Context, acc Account) (*Account, error) {
	existing, ok := m.accounts[acc.ID]
	if !ok || existing.UserID != acc.UserID || existing.DeletedAt != nil {
		return nil, ErrAccountNotFound
	}
	for _, a := range m.accounts {
		if a.ID != acc.ID && a.UserID == acc.UserID && strings.EqualFold(a.Name, acc.Name) && a.DeletedAt == nil {
			return nil, ErrAccountAlreadyExists
		}
	}
	existing.Name = acc.Name
	existing.Type = acc.Type
	existing.Status = acc.Status
	existing.UpdatedAt = time.Now()
	m.accounts[acc.ID] = existing
	return &existing, nil
}

func (m *mockAccountRepo) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	existing, ok := m.accounts[id]
	if !ok || existing.UserID != userID || existing.DeletedAt != nil {
		return ErrAccountNotFound
	}
	now := time.Now()
	existing.DeletedAt = &now
	m.accounts[id] = existing
	return nil
}

func TestAccountService_Create_Validation(t *testing.T) {
	repo := newMockAccountRepo()
	svc := NewService(repo)
	userID := uuid.New()

	// Empty name
	_, err := svc.CreateAccount(context.Background(), userID, CreateAccountInput{
		Name:           "",
		Type:           TypeBank,
		OpeningBalance: 0,
	})
	if err != ErrInvalidAccountName {
		t.Fatalf("expected ErrInvalidAccountName, got %v", err)
	}

	// Invalid type
	_, err = svc.CreateAccount(context.Background(), userID, CreateAccountInput{
		Name:           "My Bank",
		Type:           Type("CRYPTO"),
		OpeningBalance: 0,
	})
	if err != ErrInvalidAccountType {
		t.Fatalf("expected ErrInvalidAccountType, got %v", err)
	}

	// Negative opening balance
	_, err = svc.CreateAccount(context.Background(), userID, CreateAccountInput{
		Name:           "My Bank",
		Type:           TypeBank,
		OpeningBalance: -5000,
	})
	if err != ErrNegativeOpeningBalance {
		t.Fatalf("expected ErrNegativeOpeningBalance, got %v", err)
	}

	// Success
	acc, err := svc.CreateAccount(context.Background(), userID, CreateAccountInput{
		Name:           "BCA Checking",
		Type:           TypeBank,
		OpeningBalance: 1000000,
	})
	if err != nil {
		t.Fatalf("unexpected error creating account: %v", err)
	}
	if acc.CurrentBalance != 1000000 {
		t.Errorf("expected current balance 1000000, got %d", acc.CurrentBalance)
	}
	if acc.Status != StatusActive {
		t.Errorf("expected status ACTIVE, got %v", acc.Status)
	}

	// Duplicate name
	_, err = svc.CreateAccount(context.Background(), userID, CreateAccountInput{
		Name:           "bca checking",
		Type:           TypeBank,
		OpeningBalance: 500000,
	})
	if err != ErrAccountAlreadyExists {
		t.Fatalf("expected ErrAccountAlreadyExists, got %v", err)
	}
}

func TestAccountService_UpdateAndDelete(t *testing.T) {
	repo := newMockAccountRepo()
	svc := NewService(repo)
	userID := uuid.New()

	acc, _ := svc.CreateAccount(context.Background(), userID, CreateAccountInput{
		Name:           "Cash Wallet",
		Type:           TypeCash,
		OpeningBalance: 200000,
	})

	// Update
	updated, err := svc.UpdateAccount(context.Background(), userID, acc.ID, UpdateAccountInput{
		Name:   "Physical Wallet",
		Type:   TypeCash,
		Status: StatusActive,
	})
	if err != nil {
		t.Fatalf("unexpected error updating account: %v", err)
	}
	if updated.Name != "Physical Wallet" {
		t.Errorf("expected updated name 'Physical Wallet', got '%s'", updated.Name)
	}

	// Delete
	err = svc.DeleteAccount(context.Background(), userID, acc.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting account: %v", err)
	}

	// Get after delete should return ErrAccountNotFound
	_, err = svc.GetAccount(context.Background(), userID, acc.ID)
	if err != ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound after soft delete, got %v", err)
	}
}
