// Package inventory implements business logic and data access for inventory items.
package inventory

// Inventory represents a single inventory item.
type Inventory struct {
	ID         string
	Name       string
	CategoryID string
	TotalStock int64
}
