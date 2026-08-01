package equipment

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
