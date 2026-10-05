package budget

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

var (
	ErrBudgetNotFound      = errors.New("budget not found")
	ErrBudgetAlreadyExists = errors.New("a budget for this category and cycle already exists")
	ErrInvalidPlannedAmount = errors.New("planned amount must be non-negative")
	ErrInvalidCycleDates   = errors.New("cycle_start must be before cycle_end and both must be valid dates (YYYY-MM-DD)")
	ErrInvalidCategoryID   = errors.New("category_id is required")
	ErrCategoryNotFound    = errors.New("specified category does not exist or does not belong to user")
)

// Budget represents a planned spending limit for a category within a financial cycle.
type Budget struct {
	ID                     uuid.UUID  `json:"id"`
	UserID                 uuid.UUID  `json:"user_id"`
	CategoryID             uuid.UUID  `json:"category_id"`
	CycleStart             string     `json:"cycle_start"`      // YYYY-MM-DD
	CycleEnd               string     `json:"cycle_end"`         // YYYY-MM-DD
	PlannedAmount          int64      `json:"planned_amount"`    // In cents
	FormattedPlannedAmount string     `json:"formatted_planned_amount,omitempty"`
	ActualAmount           int64      `json:"actual_amount"`     // In cents
	FormattedActualAmount  string     `json:"formatted_actual_amount,omitempty"`
	Variance               int64      `json:"variance"`          // In cents
	FormattedVariance      string     `json:"formatted_variance,omitempty"`
	CategoryName           string     `json:"category_name,omitempty"`
	CategoryIcon           string     `json:"category_icon,omitempty"`
	CategoryColor          string     `json:"category_color,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty"`
}

// CreateBudgetInput defines the payload for creating a budget.
type CreateBudgetInput struct {
	CategoryID    uuid.UUID   `json:"category_id"`
	CycleStart    string      `json:"cycle_start"`    // YYYY-MM-DD
	CycleEnd      string      `json:"cycle_end"`      // YYYY-MM-DD
	PlannedAmount money.Cents `json:"planned_amount"` // In cents
}

// UpdateBudgetInput defines the payload for updating a budget.
type UpdateBudgetInput struct {
	PlannedAmount money.Cents `json:"planned_amount"` // In cents
}
