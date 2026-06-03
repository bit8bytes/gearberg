package manufacturers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	genmfr "github.com/bit8bytes/gearberg/database/queries/gen/manufacturers"
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

// GetByID returns the manufacturer with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*Manufacturer, error) {
	row, err := r.manufacturers.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
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
		return nil, fmt.Errorf("Create: %w", database.NormalizeError(err))
	}
	m := Manufacturer{
		ID:    row.ID,
		OrgID: row.OrgID,
		Name:  row.Name,
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
		return nil, fmt.Errorf("Update: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Delete removes the manufacturer. Returns database.ErrForeignKeyViolation when inventory
// items are still assigned to the manufacturer.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.manufacturers.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

func toModel(row genmfr.Manufacturer) Manufacturer {
	return Manufacturer{
		ID:    row.ID,
		OrgID: row.OrgID,
		Name:  row.Name,
	}
}
