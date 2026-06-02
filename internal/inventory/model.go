// Package inventory implements business logic and data access for inventory items.
package inventory

import "fmt"

// PurchasePriceInput formats the purchase price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (inv *Inventory) PurchasePriceInput() string { return priceInput(inv.PurchasePrice) }

// RentalPriceInput formats the rental price as a plain decimal string for use in
// form inputs (e.g. 1999 → "19.99"). Returns "" when nil.
func (inv *Inventory) RentalPriceInput() string { return priceInput(inv.RentalPrice) }

// priceInput formats cents as "%.2f" using dot separator. Form parsers accept both
// dot and comma, so dot is always safe as a display format for editable inputs.
func priceInput(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(*v)/100)
}

// Inventory represents a single inventory item.
type Inventory struct {
	ID              string
	OrgID           string
	TypeID          int64
	UsageTypeID     int64
	Name            string
	Code            string
	CategoryID      string
	CategoryName    string
	ManufacturerID  string
	StorageObjectID *string
	ImageURL        string
	TotalStock      int64
	PurchasePrice   *int64
	RentalPrice     *int64
	Notes           string
	CreatedAt       int64
	UpdatedAt       int64
}

// Base holds fields shared between bulk and serialized creation.
type Base struct {
	ID            string
	OrgID         string
	UsageTypeID   int64
	Name          string
	CategoryID    string
	Code          string
	PurchasePrice *int64
	RentalPrice   *int64
	Notes         string
}

// CreateBulkInventory holds the data required to create a bulk inventory item.
type CreateBulkInventory struct {
	Base
	TotalStock int64
}

// CreateUnit holds the data required to create a single serialized inventory unit.
type CreateUnit struct {
	ID               string
	InventoryID      string
	UnitNumber       int64
	SerialNumber     string
	NextInspectionAt *int64
}

// CreateSerializedInventory holds the data required to create a serialized inventory item with units.
type CreateSerializedInventory struct {
	Base
	Units []CreateUnit
}

// Unit represents a single serialized unit.
type Unit struct {
	ID               string
	InventoryID      string
	StatusID         int64
	UnitNumber       int64
	SerialNumber     string
	Notes            string
	NextInspectionAt *int64
	CreatedAt        int64
	UpdatedAt        int64
}

// UpdateInventory holds the data required to update an inventory item.
type UpdateInventory struct {
	ID            string
	Name          string
	CategoryID    string
	Code          string
	TotalStock    int64
	PurchasePrice *int64
	RentalPrice   *int64
	Notes         string
}

// SetImage links or unlinks a storage object from an inventory item.
type SetImage struct {
	ID              string
	StorageObjectID *string
}

// AddUnit holds the data required to add a single unit to a serialized inventory item.
type AddUnit struct {
	ID          string
	InventoryID string
	UnitNumber  int64
}

// UpdateUnit holds the data required to update a single unit's editable fields.
type UpdateUnit struct {
	ID               string
	SerialNumber     string
	Notes            string
	NextInspectionAt *int64
}
