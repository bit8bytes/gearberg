// Package inventory implements business logic and data access for inventory items.
package inventory

// Inventory represents a single inventory item.
type Inventory struct {
	ID              string
	OrgID           string
	Name            string
	CategoryID      string
	CategoryName    string
	ManufacturerID  string
	StorageObjectID *string
	ImageURL        string
	TotalStock      int64
	PurchasePrice   *float64
	RentalPrice     *float64
	Notes           string
	CreatedAt       int64
	UpdatedAt       int64
}

// CreateInventory holds the data required to create a new inventory item.
type CreateInventory struct {
	ID            string
	OrgID         string
	Name          string
	CategoryID    string
	TotalStock    int64
	PurchasePrice *float64
	RentalPrice   *float64
	Notes         string
}

// UpdateInventory holds the data required to update an inventory item.
type UpdateInventory struct {
	ID            string
	Name          string
	CategoryID    string
	TotalStock    int64
	PurchasePrice *float64
	RentalPrice   *float64
	Notes         string
}

// SetImage links or unlinks a storage object from an inventory item.
type SetImage struct {
	ID              string
	StorageObjectID *string
}
