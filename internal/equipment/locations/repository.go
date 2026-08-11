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
package locations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genloc "github.com/bit8bytes/gearberg/internal/database/queries/gen/warehouselocations"
)

// Repository provides data access for locations.
type Repository struct {
	db        *sql.DB
	locations genloc.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:        db,
		locations: genloc.New(db),
	}
}

// Count returns the number of locations belonging to orgID.
func (r *Repository) Count(ctx context.Context, orgID string) (int64, error) {
	n, err := r.locations.CountByOrgID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// GetByOrgID returns all locations belonging to orgID, ordered by name.
func (r *Repository) GetByOrgID(ctx context.Context, orgID string) ([]Location, error) {
	rows, err := r.locations.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("GetByOrgID: %w", err)
	}
	result := make([]Location, len(rows))
	for i, row := range rows {
		result[i] = toModel(row)
	}
	return result, nil
}

// GetByID returns the location with id, or ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*Location, error) {
	row, err := r.locations.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// GetByName returns the location with the given name within orgID, or ErrNotFound.
func (r *Repository) GetByName(ctx context.Context, orgID, name string) (*Location, error) {
	row, err := r.locations.GetByName(ctx, genloc.GetByNameParams{OrgID: orgID, Name: name})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Create inserts a new location.
func (r *Repository) Create(ctx context.Context, c CreateLocation) (*Location, error) {
	var parentID sql.NullString
	if c.ParentWarehouseLocationID != nil {
		parentID = sql.NullString{String: *c.ParentWarehouseLocationID, Valid: true}
	}
	row, err := r.locations.Create(ctx, genloc.CreateParams{
		ID:                        c.ID,
		ParentWarehouseLocationID: parentID,
		OrgID:                     c.OrgID,
		Name:                      c.Name,
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
	m := toModel(row)
	return &m, nil
}

// Update updates the name of the location identified by u.ID.
func (r *Repository) Update(ctx context.Context, u UpdateLocation) (*Location, error) {
	row, err := r.locations.Update(ctx, genloc.UpdateParams{
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

// Delete removes the location. Equipment items referencing it will have their location_id set to NULL.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.locations.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

func toModel(row genloc.WarehouseLocation) Location {
	var parentID *string
	if row.ParentWarehouseLocationID.Valid {
		parentID = &row.ParentWarehouseLocationID.String
	}
	return Location{
		ID:                        row.ID,
		ParentWarehouseLocationID: parentID,
		OrgID:                     row.OrgID,
		Name:                      row.Name,
	}
}
