// Package categories handles equipment category routes, business logic, and data access.
package categories

// DefaultColor is the color assigned to a category when none is specified.
const DefaultColor = "#6366f1"

// EquipmentCategory represents an equipment category belonging to a org.
type EquipmentCategory struct {
	ID    string
	OrgID string
	Name  string
	Color string
}

// CreateEquipmentCategory holds the data required to create a new equipment category.
type CreateEquipmentCategory struct {
	ID    string
	OrgID string
	Name  string
	Color string
}

// UpdateEquipmentCategory holds the data required to update an equipment category.
type UpdateEquipmentCategory struct {
	ID    string
	Name  string
	Color string
}
