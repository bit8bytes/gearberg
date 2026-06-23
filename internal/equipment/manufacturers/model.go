// Package manufacturers handles manufacturer routes, business logic, and data access.
package manufacturers

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
