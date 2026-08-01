// Package manufacturers handles manufacturer routes, business logic, and data access.
package manufacturers

import "errors"

var (
	// ErrNotFound is returned when a manufacturer does not exist.
	ErrNotFound = errors.New("manufacturer not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("manufacturer already exists")
	// ErrLimitExceeded is returned when an org's manufacturer limit is reached.
	ErrLimitExceeded = errors.New("manufacturer limit exceeded")
	// ErrInUse is returned when a manufacturer cannot be deleted due to assigned inventory items.
	ErrInUse = errors.New("manufacturer is in use and cannot be deleted")
)

// Manufacturer represents a manufacturer belonging to an org.
type Manufacturer struct {
	ID        string
	OrgID     string
	Name      string
	UpdatedAt int64
	CreatedAt int64
}

// CreateManufacturer holds the data required to create a new manufacturer.
type CreateManufacturer struct {
	ID    string
	OrgID string
	Name  string
}

// UpdateManufacturer holds the data required to update a manufacturer.
type UpdateManufacturer struct {
	ID   string
	Name string
}
