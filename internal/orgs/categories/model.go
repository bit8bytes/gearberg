// Package categories handles equipment category routes, business logic, and data access.
package categories

// EquipmentCategory represents an equipment category belonging to a org.
type EquipmentCategory struct {
	ID    string
	OrgID string
	Name  string
}

// CreateEquipmentCategory holds the data required to create a new equipment category.
type CreateEquipmentCategory struct {
	ID    string
	OrgID string
	Name  string
}

// UpdateEquipmentCategory holds the data required to update an equipment category.
type UpdateEquipmentCategory struct {
	ID   string
	Name string
}
