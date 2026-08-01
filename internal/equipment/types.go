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
