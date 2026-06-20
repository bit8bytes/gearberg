package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/equipment"
)

// seedReferenceData populates fixed lookup tables on every startup using
// INSERT OR IGNORE, so it is safe to run repeatedly.
//
// These tables hold values that are defined in application code — adding or
// renaming a value only requires changing the relevant package constant and
// this file, with no migration needed.
func seedReferenceData(ctx context.Context, db *sql.DB) error {
	for _, t := range []equipment.Type{
		equipment.Bulk,
		equipment.Serialized,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO tracking_types (id, name) VALUES (?, ?)`, t.ID(), t.String()); err != nil {
			return fmt.Errorf("seed tracking_types: %w", err)
		}
	}

	for _, u := range []equipment.UsageType{
		equipment.Rental,
		equipment.Sale,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO usage_types (id, name) VALUES (?, ?)`, u.ID(), u.String()); err != nil {
			return fmt.Errorf("seed usage_types: %w", err)
		}
	}

	return nil
}
