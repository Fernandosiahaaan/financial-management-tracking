package receivable

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Repository defines data access operations for receivables and repayments.
type Repository interface {
	Create(ctx context.Context, rec Receivable) (*Receivable, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Receivable, error)
	List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Receivable, error)
	RecordPayment(ctx context.Context, payment ReceivablePayment) (*ReceivablePayment, *Receivable, error)
	UpdateStatus(ctx context.Context, userID, id uuid.UUID, status Status, notes *string) (*Receivable, error)
	SoftDelete(ctx context.Context, userID, id uuid.UUID) error
	GetTotalActiveReceivables(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

func formatReceivable(r *Receivable) {
	if r == nil {
		return
	}
	r.FormattedPrincipal = money.FormatIDR(r.Principal)
	r.FormattedTotalPaid = money.FormatIDR(r.TotalPaid)
	r.FormattedRemainingAmount = money.FormatIDR(r.RemainingAmount)
	for i := range r.Payments {
		r.Payments[i].FormattedAmount = money.FormatIDR(r.Payments[i].Amount)
	}
}

// NewRepository creates a new PostgresRepository.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create inserts a new receivable and, if a source account is specified, atomically deducts the principal.
func (r *PostgresRepository) Create(ctx context.Context, rec Receivable) (*Receivable, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var sourceAccName string
	if rec.SourceAccountID != nil {
		// Verify account exists, belongs to user, and has sufficient funds
		var currentBalance int64
		err := tx.QueryRow(ctx, `
			SELECT name, current_balance
			FROM accounts
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
			FOR UPDATE
		`, *rec.SourceAccountID, rec.UserID).Scan(&sourceAccName, &currentBalance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrAccountNotFound
			}
			return nil, fmt.Errorf("lookup source account: %w", err)
		}

		if currentBalance < rec.Principal {
			return nil, ErrInsufficientBalance
		}

		// Deduct principal from source account
		_, err = tx.Exec(ctx, `
			UPDATE accounts
			SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, rec.Principal, *rec.SourceAccountID, rec.UserID)
		if err != nil {
			return nil, fmt.Errorf("deduct source account balance: %w", err)
		}
	}

	query := `
		INSERT INTO receivables (
			user_id, source_account_id, counterparty, principal, status, due_date, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING id, created_at, updated_at;
	`

	var created Receivable
	created = rec
	created.SourceAccountName = sourceAccName
	created.TotalPaid = 0
	created.RemainingAmount = rec.Principal

	err = tx.QueryRow(ctx, query,
		rec.UserID,
		rec.SourceAccountID,
		rec.Counterparty,
		rec.Principal,
		rec.Status,
		rec.DueDate,
		rec.Notes,
	).Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert receivable: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &created, nil
}

// GetByID fetches a receivable by ID with payment history.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Receivable, error) {
	query := `
		SELECT
			r.id, r.user_id, r.source_account_id, COALESCE(a.name, ''),
			r.counterparty, r.principal, r.status,
			TO_CHAR(r.due_date, 'YYYY-MM-DD'), r.notes,
			r.created_at, r.updated_at,
			COALESCE(p.total_paid, 0)
		FROM receivables r
		LEFT JOIN accounts a ON a.id = r.source_account_id
		LEFT JOIN (
			SELECT receivable_id, SUM(amount) AS total_paid
			FROM receivable_payments
			GROUP BY receivable_id
		) p ON p.receivable_id = r.id
		WHERE r.id = $1 AND r.user_id = $2 AND r.deleted_at IS NULL;
	`

	var rec Receivable
	var dueDateVal *string
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.SourceAccountID,
		&rec.SourceAccountName,
		&rec.Counterparty,
		&rec.Principal,
		&rec.Status,
		&dueDateVal,
		&rec.Notes,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&rec.TotalPaid,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReceivableNotFound
		}
		return nil, fmt.Errorf("query receivable: %w", err)
	}
	rec.DueDate = dueDateVal
	rec.RemainingAmount = rec.Principal - rec.TotalPaid
	if rec.RemainingAmount < 0 {
		rec.RemainingAmount = 0
	}

	// Fetch payments
	paymentsQuery := `
		SELECT
			p.id, p.receivable_id, p.user_id, p.target_account_id,
			COALESCE(a.name, ''), p.amount, TO_CHAR(p.payment_date, 'YYYY-MM-DD'),
			p.notes, p.created_at, p.updated_at
		FROM receivable_payments p
		LEFT JOIN accounts a ON a.id = p.target_account_id
		WHERE p.receivable_id = $1 AND p.user_id = $2
		ORDER BY p.payment_date DESC, p.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, paymentsQuery, id, userID)
	if err != nil {
		return nil, fmt.Errorf("query receivable payments: %w", err)
	}
	defer rows.Close()

	var payments []ReceivablePayment
	for rows.Next() {
		var p ReceivablePayment
		if err := rows.Scan(
			&p.ID,
			&p.ReceivableID,
			&p.UserID,
			&p.TargetAccountID,
			&p.TargetAccountName,
			&p.Amount,
			&p.PaymentDate,
			&p.Notes,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan receivable payment: %w", err)
		}
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []ReceivablePayment{}
	}
	rec.Payments = payments
	formatReceivable(&rec)

	return &rec, nil
}

// List returns user receivables matching filters with aggregated payments.
func (r *PostgresRepository) List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Receivable, error) {
	query := `
		SELECT
			r.id, r.user_id, r.source_account_id, COALESCE(a.name, ''),
			r.counterparty, r.principal, r.status,
			TO_CHAR(r.due_date, 'YYYY-MM-DD'), r.notes,
			r.created_at, r.updated_at,
			COALESCE(p.total_paid, 0)
		FROM receivables r
		LEFT JOIN accounts a ON a.id = r.source_account_id
		LEFT JOIN (
			SELECT receivable_id, SUM(amount) AS total_paid
			FROM receivable_payments
			GROUP BY receivable_id
		) p ON p.receivable_id = r.id
		WHERE r.user_id = $1 AND r.deleted_at IS NULL
	`
	args := []interface{}{userID}

	if filter.Status != nil {
		args = append(args, *filter.Status)
		query += fmt.Sprintf(" AND r.status = $%d", len(args))
	}

	query += " ORDER BY r.created_at DESC;"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list receivables: %w", err)
	}
	defer rows.Close()

	var receivables []Receivable
	for rows.Next() {
		var rec Receivable
		var dueDateVal *string
		err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&rec.SourceAccountID,
			&rec.SourceAccountName,
			&rec.Counterparty,
			&rec.Principal,
			&rec.Status,
			&dueDateVal,
			&rec.Notes,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.TotalPaid,
		)
		if err != nil {
			return nil, fmt.Errorf("scan receivable: %w", err)
		}
		rec.DueDate = dueDateVal
		rec.RemainingAmount = rec.Principal - rec.TotalPaid
		if rec.RemainingAmount < 0 {
			rec.RemainingAmount = 0
		}
		receivables = append(receivables, rec)
	}

	if receivables == nil {
		receivables = []Receivable{}
	}
	for i := range receivables {
		formatReceivable(&receivables[i])
	}

	return receivables, nil
}

// RecordPayment atomically records a payment, credits the target account, and updates receivable status.
func (r *PostgresRepository) RecordPayment(ctx context.Context, payment ReceivablePayment) (*ReceivablePayment, *Receivable, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock and fetch receivable
	var principal int64
	var currentStatus Status
	err = tx.QueryRow(ctx, `
		SELECT principal, status
		FROM receivables
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`, payment.ReceivableID, payment.UserID).Scan(&principal, &currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrReceivableNotFound
		}
		return nil, nil, fmt.Errorf("lookup receivable for payment: %w", err)
	}

	if currentStatus == StatusPaid || currentStatus == StatusWrittenOff {
		return nil, nil, ErrAlreadySettled
	}

	// 2. Compute current total paid
	var currentTotalPaid int64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM receivable_payments
		WHERE receivable_id = $1
	`, payment.ReceivableID).Scan(&currentTotalPaid)
	if err != nil {
		return nil, nil, fmt.Errorf("compute current payments: %w", err)
	}

	remaining := principal - currentTotalPaid
	if payment.Amount > remaining {
		return nil, nil, ErrOverpayment
	}

	// 3. Verify target account and credit balance
	var targetAccName string
	res, err := tx.Exec(ctx, `
		UPDATE accounts
		SET current_balance = current_balance + $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`, payment.Amount, payment.TargetAccountID, payment.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("credit target account: %w", err)
	}
	if res.RowsAffected() == 0 {
		return nil, nil, ErrAccountNotFound
	}

	err = tx.QueryRow(ctx, "SELECT name FROM accounts WHERE id = $1", payment.TargetAccountID).Scan(&targetAccName)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch target account name: %w", err)
	}

	// 4. Insert receivable_payment record
	var createdPayment ReceivablePayment
	createdPayment = payment
	createdPayment.TargetAccountName = targetAccName

	err = tx.QueryRow(ctx, `
		INSERT INTO receivable_payments (
			receivable_id, user_id, target_account_id, amount, payment_date, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id, created_at, updated_at
	`, payment.ReceivableID, payment.UserID, payment.TargetAccountID, payment.Amount, payment.PaymentDate, payment.Notes).
		Scan(&createdPayment.ID, &createdPayment.CreatedAt, &createdPayment.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("insert payment: %w", err)
	}

	// 5. Update receivable status
	newTotalPaid := currentTotalPaid + payment.Amount
	newRemaining := principal - newTotalPaid
	newStatus := StatusPartiallyPaid
	if newRemaining == 0 {
		newStatus = StatusPaid
	}

	_, err = tx.Exec(ctx, `
		UPDATE receivables
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, newStatus, payment.ReceivableID)
	if err != nil {
		return nil, nil, fmt.Errorf("update receivable status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit payment transaction: %w", err)
	}

	// Fetch updated receivable representation
	updatedRec, err := r.GetByID(ctx, payment.UserID, payment.ReceivableID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch updated receivable: %w", err)
	}

	createdPayment.FormattedAmount = money.FormatIDR(createdPayment.Amount)

	return &createdPayment, updatedRec, nil
}

// UpdateStatus modifies the status of a receivable (e.g. marking as WRITTEN_OFF).
func (r *PostgresRepository) UpdateStatus(ctx context.Context, userID, id uuid.UUID, status Status, notes *string) (*Receivable, error) {
	query := `
		UPDATE receivables
		SET status = $1,
		    notes = CASE WHEN $2::text IS NOT NULL THEN $2 ELSE notes END,
		    updated_at = NOW()
		WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL
		RETURNING id;
	`
	var returnedID uuid.UUID
	err := r.pool.QueryRow(ctx, query, status, notes, id, userID).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReceivableNotFound
		}
		return nil, fmt.Errorf("update receivable status: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

// SoftDelete marks a receivable as deleted.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE receivables
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete receivable: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrReceivableNotFound
	}
	return nil
}

// GetTotalActiveReceivables computes total unpaid principal across all non-settled receivables.
func (r *PostgresRepository) GetTotalActiveReceivables(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(r.principal - COALESCE(p.paid, 0)), 0)
		FROM receivables r
		LEFT JOIN (
			SELECT receivable_id, SUM(amount) AS paid
			FROM receivable_payments
			GROUP BY receivable_id
		) p ON p.receivable_id = r.id
		WHERE r.user_id = $1
		  AND r.deleted_at IS NULL
		  AND r.status NOT IN ('PAID', 'WRITTEN_OFF');
	`
	var total int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("compute total active receivables: %w", err)
	}
	if total < 0 {
		total = 0
	}
	return total, nil
}
