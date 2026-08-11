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

// Package manufacturers provides manufacturers functionality.
package manufacturers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genmfr "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentmanufacturers"
)

// Repository provides data access for manufacturers.
type Repository struct {
	db            *sql.DB
	manufacturers genmfr.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:            db,
		manufacturers: genmfr.New(db),
	}
}

// Count returns the number of manufacturers belonging to orgID.
func (r *Repository) Count(ctx context.Context, orgID string) (int64, error) {
	n, err := r.manufacturers.CountByOrgID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// GetByOrgID returns all manufacturers belonging to orgID.
func (r *Repository) GetByOrgID(ctx context.Context, orgID string) ([]Manufacturer, error) {
	rows, err := r.manufacturers.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("GetByOrgID: %w", err)
	}
	result := make([]Manufacturer, len(rows))
	for i, row := range rows {
		result[i] = toModel(row)
	}
	return result, nil
}

// GetByID returns the manufacturer with id, or ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*Manufacturer, error) {
	row, err := r.manufacturers.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Create inserts a new manufacturer.
func (r *Repository) Create(ctx context.Context, c CreateManufacturer) (*Manufacturer, error) {
	row, err := r.manufacturers.Create(ctx, genmfr.CreateParams{
		ID:    c.ID,
		OrgID: c.OrgID,
		Name:  c.Name,
	})
	if err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return nil, ErrConflict
		}
		if errors.Is(normalized, database.ErrLimitExceeded) {
			return nil, ErrLimitExceeded
		}
		return nil, fmt.Errorf("Create: %w", normalized)
	}
	m := Manufacturer{
		ID:        row.ID,
		OrgID:     row.OrgID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}
	return &m, nil
}

// Update updates the name of the manufacturer identified by u.ID.
func (r *Repository) Update(ctx context.Context, u UpdateManufacturer) (*Manufacturer, error) {
	row, err := r.manufacturers.Update(ctx, genmfr.UpdateParams{
		ID:   u.ID,
		Name: u.Name,
	})
	if err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("Update: %w", normalized)
	}
	m := toModel(row)
	return &m, nil
}

// GetByName returns the manufacturer with the given name within orgID, or ErrNotFound
// when it does not exist.
func (r *Repository) GetByName(ctx context.Context, orgID, name string) (*Manufacturer, error) {
	row, err := r.manufacturers.GetByName(ctx, genmfr.GetByNameParams{
		OrgID: orgID,
		Name:  name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Delete removes the manufacturer. Returns ErrInUse when inventory items are still assigned.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.manufacturers.Delete(ctx, id); err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrForeignKeyViolation) {
			return ErrInUse
		}
		return fmt.Errorf("Delete: %w", normalized)
	}
	return nil
}

func toModel(row genmfr.EquipmentManufacturer) Manufacturer {
	return Manufacturer{
		ID:        row.ID,
		OrgID:     row.OrgID,
		Name:      row.Name,
		UpdatedAt: row.UpdatedAt,
		CreatedAt: row.CreatedAt,
	}
}
