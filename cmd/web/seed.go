package main

import (
	"context"
	"database/sql"
	"fmt"

	inventorytypes "github.com/bit8bytes/gearberg/internal/inventory/types"
)

// seedReferenceData populates fixed lookup tables on every startup using
// INSERT OR IGNORE, so it is safe to run repeatedly.
//
// These tables hold values that are defined in application code — adding or
// renaming a value only requires changing the relevant package constant and
// this file, with no migration needed.
func seedReferenceData(ctx context.Context, db *sql.DB) error {
	for _, t := range []inventorytypes.Kind{
		inventorytypes.Bulk,
		inventorytypes.Serialized,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO inventory_types (id, name) VALUES (?, ?)`, t.ID(), t.String()); err != nil {
			return fmt.Errorf("seed inventory_types: %w", err)
		}
	}

	for _, u := range []inventorytypes.Usage{
		inventorytypes.Rental,
		inventorytypes.Sale,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO usage_types (id, name) VALUES (?, ?)`, u.ID(), u.String()); err != nil {
			return fmt.Errorf("seed usage_types: %w", err)
		}
	}

	unitStatuses := []struct {
		id   inventorytypes.UnitStatus
		name string
	}{
		{inventorytypes.UnitAvailable, "available"},
		{inventorytypes.UnitInService, "in_service"},
		{inventorytypes.UnitRetired, "retired"},
	}
	for _, s := range unitStatuses {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO unit_statuses (id, name) VALUES (?, ?)`, int64(s.id), s.name); err != nil {
			return fmt.Errorf("seed unit_statuses: %w", err)
		}
	}

	return nil
}
