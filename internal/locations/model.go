// Package locations handles location routes, business logic, and data access.
package locations

// Location represents a storage location belonging to an org.
type Location struct {
	ID        string
	OrgID     string
	Name      string
	UpdatedAt int64
	CreatedAt int64
}

// CreateLocation holds the data required to create a new location.
type CreateLocation struct {
	ID    string
	OrgID string
	Name  string
}

// UpdateLocation holds the data required to update a location.
type UpdateLocation struct {
	ID   string
	Name string
}
