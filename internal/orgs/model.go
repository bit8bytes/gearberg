package orgs

import "errors"

var (
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("org already exists")
	// ErrLimitExceeded is returned when the org limit is reached.
	ErrLimitExceeded = errors.New("org limit exceeded")
)

// Org represents a org entity.
type Org struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}
