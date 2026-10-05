package receivable

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockRepository struct {
	receivables []Receivable
	payments    []ReceivablePayment
	createErr   error
	getErr      error
	paymentErr  error
	updateErr   error
	deleteErr   error
	totalRec    int64
}

func (m *mockRepository) Create(_ context.Context, rec Receivable) (*Receivable, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	rec.ID = uuid.New()
	rec.CreatedAt = time.Now()
	rec.UpdatedAt = time.Now()
	rec.RemainingAmount = rec.Principal
	m.receivables = append(m.receivables, rec)
	return &rec, nil
}

func (m *mockRepository) GetByID(_ context.Context, userID, id uuid.UUID) (*Receivable, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, r := range m.receivables {
		if r.ID == id && r.UserID == userID {
			return &r, nil
		}
	}
	return nil, ErrReceivableNotFound
}

func (m *mockRepository) List(_ context.Context, userID uuid.UUID, filter ListFilter) ([]Receivable, error) {
	var result []Receivable
	for _, r := range m.receivables {
		if r.UserID == userID {
			if filter.Status != nil && r.Status != *filter.Status {
				continue
			}
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRepository) RecordPayment(_ context.Context, payment ReceivablePayment) (*ReceivablePayment, *Receivable, error) {
	if m.paymentErr != nil {
		return nil, nil, m.paymentErr
	}
	payment.ID = uuid.New()
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	m.payments = append(m.payments, payment)

	for i, r := range m.receivables {
		if r.ID == payment.ReceivableID {
			m.receivables[i].TotalPaid += payment.Amount
			m.receivables[i].RemainingAmount = m.receivables[i].Principal - m.receivables[i].TotalPaid
			if m.receivables[i].RemainingAmount == 0 {
				m.receivables[i].Status = StatusPaid
			} else {
				m.receivables[i].Status = StatusPartiallyPaid
			}
			return &payment, &m.receivables[i], nil
		}
	}
	return nil, nil, ErrReceivableNotFound
}

func (m *mockRepository) UpdateStatus(_ context.Context, userID, id uuid.UUID, status Status, notes *string) (*Receivable, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	for i, r := range m.receivables {
		if r.ID == id && r.UserID == userID {
			m.receivables[i].Status = status
			if notes != nil {
				m.receivables[i].Notes = *notes
			}
			return &m.receivables[i], nil
		}
	}
	return nil, ErrReceivableNotFound
}

func (m *mockRepository) SoftDelete(_ context.Context, userID, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, r := range m.receivables {
		if r.ID == id && r.UserID == userID {
			now := time.Now()
			m.receivables[i].DeletedAt = &now
			return nil
		}
	}
	return ErrReceivableNotFound
}

func (m *mockRepository) GetTotalActiveReceivables(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.totalRec, nil
}

func TestReceivableService_Create(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	userID := uuid.New()

	// Empty counterparty
	_, err := svc.Create(context.Background(), userID, CreateReceivableRequest{
		Counterparty: "   ",
		Principal:    1000000,
	})
	if err != ErrInvalidCounterparty {
		t.Fatalf("expected ErrInvalidCounterparty, got %v", err)
	}

	// Non-positive principal
	_, err = svc.Create(context.Background(), userID, CreateReceivableRequest{
		Counterparty: "Budi",
		Principal:    0,
	})
	if err != ErrInvalidPrincipal {
		t.Fatalf("expected ErrInvalidPrincipal, got %v", err)
	}

	// Invalid due date format
	invalidDate := "not-a-date"
	_, err = svc.Create(context.Background(), userID, CreateReceivableRequest{
		Counterparty: "Budi",
		Principal:    1000000,
		DueDate:      &invalidDate,
	})
	if err != ErrInvalidDueDate {
		t.Fatalf("expected ErrInvalidDueDate, got %v", err)
	}

	// Valid creation
	validDate := "2026-11-01"
	rec, err := svc.Create(context.Background(), userID, CreateReceivableRequest{
		Counterparty: "Budi",
		Principal:    2500000,
		DueDate:      &validDate,
		Notes:        "Emergency loan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Counterparty != "Budi" || rec.Principal != 2500000 || rec.Status != StatusActive {
		t.Errorf("unexpected receivable fields: %+v", rec)
	}
}

func TestReceivableService_RecordPayment(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	userID := uuid.New()
	targetAcc := uuid.New()

	rec, _ := svc.Create(context.Background(), userID, CreateReceivableRequest{
		Counterparty: "Siti",
		Principal:    1000000,
	})

	// Invalid amount
	_, _, err := svc.RecordPayment(context.Background(), userID, rec.ID, RecordPaymentRequest{
		Amount:          0,
		PaymentDate:     "2026-10-15",
		TargetAccountID: targetAcc,
	})
	if err != ErrInvalidPaymentAmount {
		t.Fatalf("expected ErrInvalidPaymentAmount, got %v", err)
	}

	// Invalid payment date
	_, _, err = svc.RecordPayment(context.Background(), userID, rec.ID, RecordPaymentRequest{
		Amount:          500000,
		PaymentDate:     "invalid-date",
		TargetAccountID: targetAcc,
	})
	if err != ErrInvalidPaymentDate {
		t.Fatalf("expected ErrInvalidPaymentDate, got %v", err)
	}

	// Missing target account
	_, _, err = svc.RecordPayment(context.Background(), userID, rec.ID, RecordPaymentRequest{
		Amount:      500000,
		PaymentDate: "2026-10-15",
	})
	if err != ErrMissingTargetAccount {
		t.Fatalf("expected ErrMissingTargetAccount, got %v", err)
	}

	// Successful partial payment
	payment, updated, err := svc.RecordPayment(context.Background(), userID, rec.ID, RecordPaymentRequest{
		Amount:          400000,
		PaymentDate:     "2026-10-15",
		TargetAccountID: targetAcc,
		Notes:           "Partial repayment",
	})
	if err != nil {
		t.Fatalf("unexpected payment error: %v", err)
	}
	if payment.Amount != 400000 || updated.Status != StatusPartiallyPaid || updated.RemainingAmount != 600000 {
		t.Errorf("unexpected updated receivable: status=%s, remaining=%d", updated.Status, updated.RemainingAmount)
	}
}

func TestReceivableService_UpdateStatus(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)
	userID := uuid.New()

	rec, _ := svc.Create(context.Background(), userID, CreateReceivableRequest{
		Counterparty: "Dewi",
		Principal:    3000000,
	})

	// Invalid status
	_, err := svc.UpdateStatus(context.Background(), userID, rec.ID, UpdateStatusRequest{
		Status: Status("INVALID"),
	})
	if err != ErrInvalidStatus {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}

	// Write off
	notes := "Cannot be contacted, written off"
	updated, err := svc.UpdateStatus(context.Background(), userID, rec.ID, UpdateStatusRequest{
		Status: StatusWrittenOff,
		Notes:  &notes,
	})
	if err != nil {
		t.Fatalf("unexpected update status error: %v", err)
	}
	if updated.Status != StatusWrittenOff || updated.Notes != notes {
		t.Errorf("expected WRITTEN_OFF status, got %s with notes %s", updated.Status, updated.Notes)
	}
}
