package units

// Status represents the status of a serialized unit.
// The integer value matches the id stored in the unit_statuses table,
// which is seeded at startup rather than via migrations.
type Status int64

// Unit status identifiers seeded into the unit_statuses table.
const (
	UnitAvailable   Status = 1
	UnitDamaged     Status = 2
	UnitUnderRepair Status = 3
	UnitRetired     Status = 4
)
