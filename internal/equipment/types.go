package equipment

// Type represents an inventory type (Bulk or Serialized).
// The integer value matches the id stored in the inventory_types table,
// which is seeded at startup rather than via migrations.
type Type int64

// Equipment type identifiers seeded into the inventory_types table.
const (
	Bulk       Type = 1
	Serialized Type = 2
)

// ID returns the database id for the inventory type.
func (k Type) ID() int64 { return int64(k) }

// String returns the name stored in the inventory_types table.
func (k Type) String() string {
	switch k {
	case Bulk:
		return "bulk"
	case Serialized:
		return "serialized"
	default:
		return ""
	}
}

// TypeFromString returns the Type matching name ("bulk" or "serialized"), or 0 when unknown.
func TypeFromString(name string) Type {
	switch name {
	case "bulk":
		return Bulk
	case "serialized":
		return Serialized
	default:
		return 0
	}
}

// Label returns the human-friendly label for the inventory type.
func (k Type) Label() string {
	switch k {
	case Bulk:
		return "Bulk"
	case Serialized:
		return "Serialized"
	default:
		return ""
	}
}

// Kind represents whether equipment is a physical item or a virtual combination.
// The integer value matches the id stored in the equipment_types table,
// which is seeded at startup rather than via migrations.
type Kind int64

// Equipment kind identifiers seeded into the equipment_types table.
const (
	Physical Kind = 1
	Virtual  Kind = 2
)

// ID returns the database id for the equipment kind.
func (k Kind) ID() int64 { return int64(k) }

// String returns the name stored in the equipment_types table.
func (k Kind) String() string {
	switch k {
	case Physical:
		return "physical"
	case Virtual:
		return "virtual"
	default:
		return ""
	}
}

// KindFromString returns the Kind matching name, or 0 when unknown.
func KindFromString(name string) Kind {
	switch name {
	case "physical":
		return Physical
	case "virtual":
		return Virtual
	default:
		return 0
	}
}

// Label returns the human-friendly label for the equipment kind.
func (k Kind) Label() string {
	switch k {
	case Physical:
		return "Physical"
	case Virtual:
		return "Virtual"
	default:
		return ""
	}
}

// UsageType represents how an inventory item is used (rental or sale).
// The integer value matches the id stored in the usage_types table,
// which is seeded at startup rather than via migrations.
type UsageType int64

// Usage type identifiers seeded into the usage_types table.
const (
	Rental UsageType = 1
	Sale   UsageType = 2
)

// ID returns the database id for the usage type.
func (u UsageType) ID() int64 { return int64(u) }

// ParseUsage returns the Usage matching name, or 0 when unknown.
func ParseUsage(name string) UsageType {
	switch name {
	case "rental":
		return Rental
	case "sale":
		return Sale
	default:
		return 0
	}
}

// Label returns the human-friendly label for the usage type.
func (u UsageType) Label() string {
	switch u {
	case Rental:
		return "Rental"
	case Sale:
		return "Sale"
	default:
		return ""
	}
}

// String returns the name stored in the usage_types table.
func (u UsageType) String() string {
	switch u {
	case Rental:
		return "rental"
	case Sale:
		return "sale"
	default:
		return ""
	}
}
