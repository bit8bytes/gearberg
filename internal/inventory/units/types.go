package units

// StatusType represents the status of a serialized unit.
// The integer value matches the id stored in the unit_statuses table,
// which is seeded at startup rather than via migrations.
type StatusType int64

// Unit status identifiers seeded into the unit_statuses table.
const (
	UnitAvailable   StatusType = 1
	UnitDamaged     StatusType = 2
	UnitUnderRepair StatusType = 3
	UnitRetired     StatusType = 4
)
