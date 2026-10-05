package category

import (
	"time"

	"github.com/google/uuid"
)

// Type represents whether a category is for income or expense.
type Type string

const (
	TypeIncome  Type = "INCOME"
	TypeExpense Type = "EXPENSE"
)

// Valid returns true if the category type is supported.
func (t Type) Valid() bool {
	switch t {
	case TypeIncome, TypeExpense:
		return true
	default:
		return false
	}
}

// Category represents a transaction classification scoped to a user.
type Category struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Name      string     `json:"name"`
	Type      Type       `json:"type"`
	Icon      string     `json:"icon"`
	Color     string     `json:"color"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
