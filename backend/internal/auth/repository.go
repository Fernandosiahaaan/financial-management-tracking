package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// Repository defines data access methods for users and their settings.
type Repository interface {
	CreateUser(ctx context.Context, user User, defaultCycleStartDay int) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, cycleStartDay int) (*UserSettings, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	SetPinHash(ctx context.Context, userID uuid.UUID, pinHash *string) error
}

// PostgresRepository implements Repository using pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgresRepository.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateUser inserts a new user and corresponding default settings in a single transaction.
func (r *PostgresRepository) CreateUser(ctx context.Context, user User, defaultCycleStartDay int) (*User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	cleanEmail := strings.ToLower(strings.TrimSpace(user.Email))

	userQuery := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at, updated_at;
	`

	var created User
	err = tx.QueryRow(ctx, userQuery, cleanEmail, user.PasswordHash).Scan(
		&created.ID,
		&created.Email,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: insert user: %w", err)
	}

	settingsQuery := `
		INSERT INTO user_settings (user_id, cycle_start_day)
		VALUES ($1, $2);
	`
	if _, err := tx.Exec(ctx, settingsQuery, created.ID, defaultCycleStartDay); err != nil {
		return nil, fmt.Errorf("create user: insert settings: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create user: commit tx: %w", err)
	}

	return &created, nil
}

// GetUserByEmail queries a user by normalized email address.
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	query := `
		SELECT id, email, password_hash, pin_hash, created_at, updated_at
		FROM users
		WHERE LOWER(email) = $1;
	`

	var u User
	err := r.pool.QueryRow(ctx, query, cleanEmail).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}

// GetUserByID queries a user by primary key UUID.
func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, pin_hash, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var u User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}

// GetUserSettings retrieves settings for a specific user.
func (r *PostgresRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*UserSettings, error) {
	query := `
		SELECT user_id, cycle_start_day, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1;
	`

	var s UserSettings
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&s.UserID,
		&s.CycleStartDay,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user settings: %w", err)
	}

	return &s, nil
}

// UpdateUserSettings updates the cycle_start_day for a user.
func (r *PostgresRepository) UpdateUserSettings(ctx context.Context, userID uuid.UUID, cycleStartDay int) (*UserSettings, error) {
	query := `
		UPDATE user_settings
		SET cycle_start_day = $2, updated_at = $3
		WHERE user_id = $1
		RETURNING user_id, cycle_start_day, created_at, updated_at;
	`

	now := time.Now()
	var s UserSettings
	err := r.pool.QueryRow(ctx, query, userID, cycleStartDay, now).Scan(
		&s.UserID,
		&s.CycleStartDay,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user settings: %w", err)
	}

	return &s, nil
}

// GetUserProfile combines user identity and cycle settings.
func (r *PostgresRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	query := `
		SELECT u.id, u.email, COALESCE(s.cycle_start_day, 1), u.created_at, (u.pin_hash IS NOT NULL AND u.pin_hash != '') AS has_pin
		FROM users u
		LEFT JOIN user_settings s ON u.id = s.user_id
		WHERE u.id = $1;
	`

	var p UserProfile
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.ID,
		&p.Email,
		&p.CycleStartDay,
		&p.CreatedAt,
		&p.HasPin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	return &p, nil
}

// SetPinHash updates or clears the pin_hash for a user.
func (r *PostgresRepository) SetPinHash(ctx context.Context, userID uuid.UUID, pinHash *string) error {
	query := `
		UPDATE users
		SET pin_hash = $2, updated_at = $3
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, userID, pinHash, time.Now())
	if err != nil {
		return fmt.Errorf("set pin hash: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

