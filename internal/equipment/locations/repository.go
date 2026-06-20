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

// GetByID returns the location with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*Location, error) {
	row, err := r.locations.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// GetByName returns the location with the given name within orgID, or database.ErrNotFound.
func (r *Repository) GetByName(ctx context.Context, orgID, name string) (*Location, error) {
	row, err := r.locations.GetByName(ctx, genloc.GetByNameParams{OrgID: orgID, Name: name})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
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
		return nil, fmt.Errorf("Create: %w", database.NormalizeError(err))
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
		return nil, fmt.Errorf("Update: %w", err)
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
