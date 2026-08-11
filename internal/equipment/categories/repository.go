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

// Package categories provides categories functionality.
package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genec "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentcategories"
)

// Repository provides data access for equipment equipmentCategories.
type Repository struct {
	db                  *sql.DB
	equipmentCategories genec.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:                  db,
		equipmentCategories: genec.New(db),
	}
}

// Count returns the number of equipment categories belonging to orgID.
func (r *Repository) Count(ctx context.Context, orgID string) (int64, error) {
	n, err := r.equipmentCategories.CountByOrgID(ctx, sql.NullString{String: orgID, Valid: orgID != ""})
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// GetByOrgID returns all equipmentCategories belonging to orgID.
func (r *Repository) GetByOrgID(ctx context.Context, orgID string) ([]EquipmentCategory, error) {
	rows, err := r.equipmentCategories.GetByOrgID(ctx, sql.NullString{String: orgID, Valid: orgID != ""})
	if err != nil {
		return nil, fmt.Errorf("GetByOrgID: %w", err)
	}
	result := make([]EquipmentCategory, len(rows))
	for i, row := range rows {
		result[i] = toModel(row)
	}
	return result, nil
}

// GetByID returns the category with id, or ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Create inserts a new equipment category.
func (r *Repository) Create(ctx context.Context, c CreateEquipmentCategory) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.Create(ctx, genec.CreateParams{
		ID:    c.ID,
		OrgID: sql.NullString{String: c.OrgID, Valid: c.OrgID != ""},
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
	m := EquipmentCategory{
		ID:        row.ID,
		OrgID:     row.OrgID.String,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}
	return &m, nil
}

// Update updates the name of the category identified by u.ID.
func (r *Repository) Update(ctx context.Context, u UpdateEquipmentCategory) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.Update(ctx, genec.UpdateParams{
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

// GetByName returns the category with the given name within orgID, or ErrNotFound
// when it does not exist.
func (r *Repository) GetByName(ctx context.Context, orgID, name string) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.GetByName(ctx, genec.GetByNameParams{
		OrgID: sql.NullString{String: orgID, Valid: orgID != ""},
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

// Delete removes the category. Returns ErrInUse when inventory items are still assigned.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.equipmentCategories.Delete(ctx, id); err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrForeignKeyViolation) {
			return ErrInUse
		}
		return fmt.Errorf("Delete: %w", normalized)
	}
	return nil
}

func toModel(row genec.EquipmentCategory) EquipmentCategory {
	return EquipmentCategory{
		ID:        row.ID,
		OrgID:     row.OrgID.String,
		Name:      row.Name,
		UpdatedAt: row.UpdatedAt,
		CreatedAt: row.CreatedAt,
	}
}
