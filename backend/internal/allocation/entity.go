package allocation

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/fernando/financial-management-tracking/backend/internal/money"
)

var (
	ErrAllocationNotFound      = errors.New("allocation not found")
	ErrAllocationAlreadyExists = errors.New("an allocation for this category and cycle already exists")
	ErrInvalidAllocatedAmount  = errors.New("allocated amount must be non-negative")
	ErrInvalidCycleDates       = errors.New("cycle_start must be before cycle_end and both must be valid dates (YYYY-MM-DD)")
	ErrInvalidCategoryID       = errors.New("category_id is required")
	ErrCategoryNotFound        = errors.New("specified category does not exist or does not belong to user")
	ErrExceedsAvailableIncome  = errors.New("total allocations exceed available income for this cycle")
)

// Allocation represents income distributed to a purpose category within a financial cycle.
type Allocation struct {
	ID                       uuid.UUID  `json:"id"`
	UserID                   uuid.UUID  `json:"user_id"`
	CategoryID               uuid.UUID  `json:"category_id"`
	CycleStart               string     `json:"cycle_start"`                 // YYYY-MM-DD
	CycleEnd                 string     `json:"cycle_end"`                   // YYYY-MM-DD
	AllocatedAmount          int64      `json:"allocated_amount"`            // In cents
	FormattedAllocatedAmount string     `json:"formatted_allocated_amount,omitempty"`
	CategoryName             string     `json:"category_name,omitempty"`
	CategoryIcon             string     `json:"category_icon,omitempty"`
	CategoryColor            string     `json:"category_color,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	DeletedAt                *time.Time `json:"deleted_at,omitempty"`
}

// AllocationSummary provides the allocation distribution overview for a cycle.
type AllocationSummary struct {
	Allocations              []Allocation `json:"allocations"`
	TotalAllocated           int64        `json:"total_allocated"`
	FormattedTotalAllocated  string       `json:"formatted_total_allocated,omitempty"`
	TotalIncome              int64        `json:"total_income"`
	FormattedTotalIncome     string       `json:"formatted_total_income,omitempty"`
	RemainingIncome          int64        `json:"remaining_income"`
	FormattedRemainingIncome string       `json:"formatted_remaining_income,omitempty"`
}

// CreateAllocationInput defines the payload for creating an allocation.
type CreateAllocationInput struct {
	CategoryID      uuid.UUID   `json:"category_id"`
	CycleStart      string      `json:"cycle_start"`      // YYYY-MM-DD
	CycleEnd        string      `json:"cycle_end"`         // YYYY-MM-DD
	AllocatedAmount money.Cents `json:"allocated_amount"` // In cents
}

// UpdateAllocationInput defines the payload for updating an allocation.
type UpdateAllocationInput struct {
	AllocatedAmount money.Cents `json:"allocated_amount"` // In cents
}
