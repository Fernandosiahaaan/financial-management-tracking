package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated account in the system.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserSettings stores personalized preferences such as the financial cycle start day.
type UserSettings struct {
	UserID        uuid.UUID `json:"user_id"`
	CycleStartDay int       `json:"cycle_start_day"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserProfile provides a unified profile view for frontend consumers.
type UserProfile struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	CycleStartDay int       `json:"cycle_start_day"`
	CreatedAt     time.Time `json:"created_at"`
}
