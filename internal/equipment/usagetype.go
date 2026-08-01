package equipment

// UsageType represents how an inventory item is used (rental or sale).
// The integer value matches the id stored in the usage_types table,
// which is seeded at startup rather than via migrations.
type UsageType int64

// Usage type identifiers seeded into the usage_types table.
const (
	Rental UsageType = 1
)

// ID returns the database id for the usage type.
func (u UsageType) ID() int64 { return int64(u) }

// ParseUsage returns the Usage matching name, or 0 when unknown.
func ParseUsage(name string) UsageType {
	switch name {
	case "rental":
		return Rental
	default:
		return 0
	}
}

// Label returns the human-friendly label for the usage type.
func (u UsageType) Label() string {
	switch u {
	case Rental:
		return "Rental"
	default:
		return ""
	}
}

// String returns the name stored in the usage_types table.
func (u UsageType) String() string {
	switch u {
	case Rental:
		return "rental"
	default:
		return ""
	}
}
