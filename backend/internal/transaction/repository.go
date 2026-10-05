package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

// Repository defines data access methods for transactions and atomic balance updates.
type Repository interface {
	Create(ctx context.Context, tx Transaction) (*Transaction, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Transaction, error)
	List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Transaction, error)
	Update(ctx context.Context, userID, id uuid.UUID, newTx Transaction) (*Transaction, error)
	SoftDelete(ctx context.Context, userID, id uuid.UUID) error
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create inserts a new transaction and applies atomic balance modifications.
func (r *PostgresRepository) Create(ctx context.Context, tx Transaction) (*Transaction, error) {
	pgxTx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer pgxTx.Rollback(ctx)

	// 1. Verify account exists and update balance
	switch tx.Type {
	case TypeIncome:
		res, err := pgxTx.Exec(ctx, `
			UPDATE accounts
			SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, tx.AccountID, tx.UserID)
		if err != nil {
			return nil, fmt.Errorf("update income balance: %w", err)
		}
		if res.RowsAffected() == 0 {
			return nil, ErrAccountNotFound
		}

	case TypeExpense:
		res, err := pgxTx.Exec(ctx, `
			UPDATE accounts
			SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, tx.AccountID, tx.UserID)
		if err != nil {
			return nil, fmt.Errorf("update expense balance: %w", err)
		}
		if res.RowsAffected() == 0 {
			return nil, ErrAccountNotFound
		}

	case TypeTransfer:
		// Source account
		res, err := pgxTx.Exec(ctx, `
			UPDATE accounts
			SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, tx.AccountID, tx.UserID)
		if err != nil {
			return nil, fmt.Errorf("update transfer source balance: %w", err)
		}
		if res.RowsAffected() == 0 {
			return nil, fmt.Errorf("source account: %w", ErrAccountNotFound)
		}

		// Destination account
		resDest, err := pgxTx.Exec(ctx, `
			UPDATE accounts
			SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, *tx.DestinationAccountID, tx.UserID)
		if err != nil {
			return nil, fmt.Errorf("update transfer dest balance: %w", err)
		}
		if resDest.RowsAffected() == 0 {
			return nil, fmt.Errorf("destination account: %w", ErrAccountNotFound)
		}
	}

	// 2. Insert transaction
	insertQuery := `
		INSERT INTO transactions (
			user_id, account_id, destination_account_id, category_id,
			type, amount, transaction_date, description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, account_id, destination_account_id, category_id,
		          type, amount, transaction_date::text, description, created_at, updated_at, deleted_at;
	`

	var created Transaction
	err = pgxTx.QueryRow(ctx, insertQuery,
		tx.UserID,
		tx.AccountID,
		tx.DestinationAccountID,
		tx.CategoryID,
		tx.Type,
		tx.Amount,
		tx.TransactionDate,
		tx.Description,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.AccountID,
		&created.DestinationAccountID,
		&created.CategoryID,
		&created.Type,
		&created.Amount,
		&created.TransactionDate,
		&created.Description,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	if err := pgxTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Populate joined names for display
	return r.GetByID(ctx, created.UserID, created.ID)
}

// GetByID returns a transaction with joined account and category details.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT t.id, t.user_id, t.account_id, t.destination_account_id, t.category_id,
		       t.type, t.amount, t.transaction_date::text, t.description,
		       t.created_at, t.updated_at, t.deleted_at,
		       a.name AS account_name,
		       COALESCE(da.name, '') AS destination_account_name,
		       COALESCE(c.name, '') AS category_name,
		       COALESCE(c.color, '') AS category_color,
		       COALESCE(c.icon, '') AS category_icon
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		LEFT JOIN accounts da ON da.id = t.destination_account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.id = $1 AND t.user_id = $2 AND t.deleted_at IS NULL;
	`

	var tx Transaction
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.AccountID,
		&tx.DestinationAccountID,
		&tx.CategoryID,
		&tx.Type,
		&tx.Amount,
		&tx.TransactionDate,
		&tx.Description,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&tx.DeletedAt,
		&tx.AccountName,
		&tx.DestinationAccountName,
		&tx.CategoryName,
		&tx.CategoryColor,
		&tx.CategoryIcon,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	tx.FormattedAmount = money.FormatIDR(tx.Amount)

	return &tx, nil
}

// List returns transactions matching the filter criteria.
func (r *PostgresRepository) List(ctx context.Context, userID uuid.UUID, filter ListFilter) ([]Transaction, error) {
	query := `
		SELECT t.id, t.user_id, t.account_id, t.destination_account_id, t.category_id,
		       t.type, t.amount, t.transaction_date::text, t.description,
		       t.created_at, t.updated_at, t.deleted_at,
		       a.name AS account_name,
		       COALESCE(da.name, '') AS destination_account_name,
		       COALESCE(c.name, '') AS category_name,
		       COALESCE(c.color, '') AS category_color,
		       COALESCE(c.icon, '') AS category_icon
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		LEFT JOIN accounts da ON da.id = t.destination_account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1 AND t.deleted_at IS NULL
	`

	args := []any{userID}
	paramIdx := 2

	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND t.transaction_date >= $%d", paramIdx)
		args = append(args, filter.StartDate)
		paramIdx++
	}

	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND t.transaction_date <= $%d", paramIdx)
		args = append(args, filter.EndDate)
		paramIdx++
	}

	if filter.Type != nil && (*filter.Type).Valid() {
		query += fmt.Sprintf(" AND t.type = $%d", paramIdx)
		args = append(args, *filter.Type)
		paramIdx++
	}

	if filter.AccountID != nil {
		query += fmt.Sprintf(" AND (t.account_id = $%d OR t.destination_account_id = $%d)", paramIdx, paramIdx)
		args = append(args, *filter.AccountID)
		paramIdx++
	}

	query += " ORDER BY t.transaction_date DESC, t.created_at DESC"

	limit := 50
	if filter.Limit > 0 && filter.Limit <= 100 {
		limit = filter.Limit
	}
	query += fmt.Sprintf(" LIMIT $%d", paramIdx)
	args = append(args, limit)
	paramIdx++

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramIdx)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var list []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.AccountID,
			&tx.DestinationAccountID,
			&tx.CategoryID,
			&tx.Type,
			&tx.Amount,
			&tx.TransactionDate,
			&tx.Description,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&tx.DeletedAt,
			&tx.AccountName,
			&tx.DestinationAccountName,
			&tx.CategoryName,
			&tx.CategoryColor,
			&tx.CategoryIcon,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		tx.FormattedAmount = money.FormatIDR(tx.Amount)
		list = append(list, tx)
	}

	if list == nil {
		list = []Transaction{}
	}

	return list, nil
}

// Update modifies an existing transaction and recalculates the balance diff atomically.
func (r *PostgresRepository) Update(ctx context.Context, userID, id uuid.UUID, newTx Transaction) (*Transaction, error) {
	pgxTx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer pgxTx.Rollback(ctx)

	// 1. Fetch current transaction with row lock
	var oldTx Transaction
	err = pgxTx.QueryRow(ctx, `
		SELECT id, user_id, account_id, destination_account_id, category_id,
		       type, amount, transaction_date::text, description, created_at, updated_at
		FROM transactions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`, id, userID).Scan(
		&oldTx.ID,
		&oldTx.UserID,
		&oldTx.AccountID,
		&oldTx.DestinationAccountID,
		&oldTx.CategoryID,
		&oldTx.Type,
		&oldTx.Amount,
		&oldTx.TransactionDate,
		&oldTx.Description,
		&oldTx.CreatedAt,
		&oldTx.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("lock existing transaction: %w", err)
	}

	// 2. Reverse old transaction effect
	switch oldTx.Type {
	case TypeIncome:
		_, err = pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, oldTx.Amount, oldTx.AccountID, userID)
	case TypeExpense:
		_, err = pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, oldTx.Amount, oldTx.AccountID, userID)
	case TypeTransfer:
		_, err = pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, oldTx.Amount, oldTx.AccountID, userID)
		if err == nil && oldTx.DestinationAccountID != nil {
			_, err = pgxTx.Exec(ctx, `
				UPDATE accounts SET current_balance = current_balance - $1, updated_at = NOW()
				WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
			`, oldTx.Amount, *oldTx.DestinationAccountID, userID)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("revert old transaction effect: %w", err)
	}

	// 3. Apply new transaction effect
	switch newTx.Type {
	case TypeIncome:
		res, err := pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, newTx.Amount, newTx.AccountID, userID)
		if err != nil {
			return nil, fmt.Errorf("apply new income effect: %w", err)
		}
		if res.RowsAffected() == 0 {
			return nil, ErrAccountNotFound
		}
	case TypeExpense:
		res, err := pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, newTx.Amount, newTx.AccountID, userID)
		if err != nil {
			return nil, fmt.Errorf("apply new expense effect: %w", err)
		}
		if res.RowsAffected() == 0 {
			return nil, ErrAccountNotFound
		}
	case TypeTransfer:
		res, err := pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, newTx.Amount, newTx.AccountID, userID)
		if err != nil {
			return nil, fmt.Errorf("apply new transfer source effect: %w", err)
		}
		if res.RowsAffected() == 0 {
			return nil, fmt.Errorf("source account: %w", ErrAccountNotFound)
		}

		resDest, err := pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, newTx.Amount, *newTx.DestinationAccountID, userID)
		if err != nil {
			return nil, fmt.Errorf("apply new transfer dest effect: %w", err)
		}
		if resDest.RowsAffected() == 0 {
			return nil, fmt.Errorf("destination account: %w", ErrAccountNotFound)
		}
	}

	// 4. Update the transaction row
	updateQuery := `
		UPDATE transactions
		SET account_id = $1, destination_account_id = $2, category_id = $3,
		    type = $4, amount = $5, transaction_date = $6, description = $7, updated_at = NOW()
		WHERE id = $8 AND user_id = $9 AND deleted_at IS NULL;
	`
	_, err = pgxTx.Exec(ctx, updateQuery,
		newTx.AccountID,
		newTx.DestinationAccountID,
		newTx.CategoryID,
		newTx.Type,
		newTx.Amount,
		newTx.TransactionDate,
		newTx.Description,
		id,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("update transaction row: %w", err)
	}

	if err := pgxTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update transaction: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

// SoftDelete marks a transaction as deleted and reverses its balance adjustments atomically.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	pgxTx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer pgxTx.Rollback(ctx)

	// 1. Fetch transaction with lock
	var tx Transaction
	err = pgxTx.QueryRow(ctx, `
		SELECT id, user_id, account_id, destination_account_id, type, amount
		FROM transactions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		FOR UPDATE
	`, id, userID).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.AccountID,
		&tx.DestinationAccountID,
		&tx.Type,
		&tx.Amount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTransactionNotFound
		}
		return fmt.Errorf("lock transaction for delete: %w", err)
	}

	// 2. Reverse effect
	switch tx.Type {
	case TypeIncome:
		_, err = pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance - $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, tx.AccountID, userID)
	case TypeExpense:
		_, err = pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, tx.AccountID, userID)
	case TypeTransfer:
		_, err = pgxTx.Exec(ctx, `
			UPDATE accounts SET current_balance = current_balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
		`, tx.Amount, tx.AccountID, userID)
		if err == nil && tx.DestinationAccountID != nil {
			_, err = pgxTx.Exec(ctx, `
				UPDATE accounts SET current_balance = current_balance - $1, updated_at = NOW()
				WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
			`, tx.Amount, *tx.DestinationAccountID, userID)
		}
	}
	if err != nil {
		return fmt.Errorf("reverse balance on delete: %w", err)
	}

	// 3. Mark deleted_at
	res, err := pgxTx.Exec(ctx, `
		UPDATE transactions
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete transaction row: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}

	if err := pgxTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit soft delete: %w", err)
	}

	return nil
}
