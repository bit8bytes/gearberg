// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/credentials"
	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/federated"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/tokens"
)

// seedReferenceData populates fixed lookup tables on every startup using
// INSERT OR IGNORE, so it is safe to run repeatedly.
//
// These tables hold values that are defined in application code -- adding or
// renaming a value only requires changing the relevant package constant and
// this file, with no migration needed.
func seedReferenceData(ctx context.Context, db *sql.DB) error { //nolint:cyclop // repetitive inserts, not branching logic
	for _, t := range []tokens.Scope{
		tokens.PasswordReset,
		tokens.EmailVerification,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO token_scopes (id, name) VALUES (?, ?)`, t.ID(), t.String()); err != nil {
			return fmt.Errorf("seed token_scopes: %w", err)
		}
	}

	for _, t := range []credentials.Type{
		credentials.PasswordType,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO credential_types (id, name) VALUES (?, ?)`, t.ID(), t.String()); err != nil {
			return fmt.Errorf("seed credential_types: %w", err)
		}
	}

	for _, r := range []orgs.Role{
		orgs.OwnerRole,
		orgs.AdminRole,
		orgs.MemberRole,
		orgs.ViewerRole,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO org_roles (id, name, rank) VALUES (?, ?, ?)`, r.ID(), r.String(), r.Rank()); err != nil {
			return fmt.Errorf("seed org_roles: %w", err)
		}
	}

	for _, t := range []equipment.TrackingType{
		equipment.Bulk,
		equipment.Serialized,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO tracking_types (id, name) VALUES (?, ?)`, t.ID(), t.String()); err != nil {
			return fmt.Errorf("seed tracking_types: %w", err)
		}
	}

	for _, k := range []equipment.Kind{
		equipment.Physical,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO equipment_types (id, name) VALUES (?, ?)`, k.ID(), k.String()); err != nil {
			return fmt.Errorf("seed equipment_types: %w", err)
		}
	}

	for _, u := range []equipment.UsageType{
		equipment.Rental,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO usage_types (id, name) VALUES (?, ?)`, u.ID(), u.String()); err != nil {
			return fmt.Errorf("seed usage_types: %w", err)
		}
	}

	for _, p := range []federated.Provider{
		federated.AuthentikProvider,
	} {
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO provider_types (id, name) VALUES (?, ?)`, p.ID(), p.String()); err != nil {
			return fmt.Errorf("seed provider_types: %w", err)
		}
	}

	return nil
}
