// Package types defines enumerated values for inventory classification,
// matching the seeded lookup tables in the database.
package types

// Kind represents an inventory type (Bulk or Serialized).
// The integer value matches the id stored in the inventory_types table,
// which is seeded at startup rather than via migrations.
type Kind int64

// Inventory type identifiers seeded into the inventory_types table.
const (
	Bulk       Kind = 1
	Serialized Kind = 2
)

// ID returns the database id for the inventory type.
func (k Kind) ID() int64 { return int64(k) }

// String returns the name stored in the inventory_types table.
func (k Kind) String() string {
	switch k {
	case Bulk:
		return "bulk"
	case Serialized:
		return "serialized"
	default:
		return ""
	}
}

// Label returns the human-friendly label for the inventory type.
func (k Kind) Label() string {
	switch k {
	case Bulk:
		return "Bulk"
	case Serialized:
		return "Serialized"
	default:
		return ""
	}
}

// Usage represents how an inventory item is used (rental or sale).
// The integer value matches the id stored in the usage_types table,
// which is seeded at startup rather than via migrations.
type Usage int64

// Usage type identifiers seeded into the usage_types table.
const (
	Rental Usage = 1
	Sale   Usage = 2
)

// ID returns the database id for the usage type.
func (u Usage) ID() int64 { return int64(u) }

// ParseUsage returns the Usage matching name, or 0 when unknown.
func ParseUsage(name string) Usage {
	switch name {
	case "rental":
		return Rental
	case "sale":
		return Sale
	default:
		return 0
	}
}

// String returns the name stored in the usage_types table.
func (u Usage) String() string {
	switch u {
	case Rental:
		return "rental"
	case Sale:
		return "sale"
	default:
		return ""
	}
}
