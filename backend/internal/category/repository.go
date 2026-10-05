package category

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("a category with this name and type already exists")
)

// Repository defines data access methods for categories.
type Repository interface {
	Create(ctx context.Context, cat Category) (*Category, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	Update(ctx context.Context, cat Category) (*Category, error)
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

// Create inserts a new category scoped to the user.
func (r *PostgresRepository) Create(ctx context.Context, cat Category) (*Category, error) {
	query := `
		INSERT INTO categories (user_id, name, type, icon, color)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, type, icon, color, deleted_at, created_at, updated_at;
	`

	cleanName := strings.TrimSpace(cat.Name)
	if cat.Icon == "" {
		cat.Icon = "tag"
	}
	if cat.Color == "" {
		cat.Color = "#6366F1"
	}

	var created Category
	err := r.pool.QueryRow(ctx, query,
		cat.UserID,
		cleanName,
		cat.Type,
		cat.Icon,
		cat.Color,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Name,
		&created.Type,
		&created.Icon,
		&created.Color,
		&created.DeletedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("create category: %w", err)
	}

	return &created, nil
}

// ListByUser returns all active categories for the given user.
func (r *PostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	query := `
		SELECT id, user_id, name, type, icon, color, deleted_at, created_at, updated_at
		FROM categories
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY type ASC, name ASC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err := rows.Scan(
			&cat.ID,
			&cat.UserID,
			&cat.Name,
			&cat.Type,
			&cat.Icon,
			&cat.Color,
			&cat.DeletedAt,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if categories == nil {
		categories = []Category{}
	}

	return categories, nil
}

// GetByID finds an active category by its ID and UserID.
func (r *PostgresRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*Category, error) {
	query := `
		SELECT id, user_id, name, type, icon, color, deleted_at, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
	`

	var cat Category
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&cat.ID,
		&cat.UserID,
		&cat.Name,
		&cat.Type,
		&cat.Icon,
		&cat.Color,
		&cat.DeletedAt,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by id: %w", err)
	}

	return &cat, nil
}

// Update modifies an existing category's name, type, icon, or color.
func (r *PostgresRepository) Update(ctx context.Context, cat Category) (*Category, error) {
	query := `
		UPDATE categories
		SET name = $3, type = $4, icon = $5, color = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, name, type, icon, color, deleted_at, created_at, updated_at;
	`

	cleanName := strings.TrimSpace(cat.Name)
	var updated Category
	err := r.pool.QueryRow(ctx, query,
		cat.ID,
		cat.UserID,
		cleanName,
		cat.Type,
		cat.Icon,
		cat.Color,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Name,
		&updated.Type,
		&updated.Icon,
		&updated.Color,
		&updated.DeletedAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("update category: %w", err)
	}

	return &updated, nil
}

// SoftDelete sets deleted_at to current timestamp.
func (r *PostgresRepository) SoftDelete(ctx context.Context, userID, id uuid.UUID) error {
	query := `
		UPDATE categories
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;
	`

	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
