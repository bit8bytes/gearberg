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
	n, err := r.equipmentCategories.CountByOrgID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// GetByOrgID returns all equipmentCategories belonging to orgID.
func (r *Repository) GetByOrgID(ctx context.Context, orgID string) ([]EquipmentCategory, error) {
	rows, err := r.equipmentCategories.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("GetByOrgID: %w", err)
	}
	result := make([]EquipmentCategory, len(rows))
	for i, row := range rows {
		result[i] = toModel(row)
	}
	return result, nil
}

// GetByID returns the category with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
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
		OrgID: c.OrgID,
		Name:  c.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("Create: %w", database.NormalizeError(err))
	}
	m := EquipmentCategory{
		ID:    row.ID,
		OrgID: row.OrgID,
		Name:  row.Name,
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
		return nil, fmt.Errorf("Update: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// GetByName returns the category with the given name within orgID, or database.ErrNotFound
// when it does not exist.
func (r *Repository) GetByName(ctx context.Context, orgID, name string) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.GetByName(ctx, genec.GetByNameParams{
		OrgID: orgID,
		Name:  name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Delete removes the category. Returns database.ErrForeignKeyViolation when inventory
// items are still assigned to the category.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.equipmentCategories.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

func toModel(row genec.EquipmentCategory) EquipmentCategory {
	return EquipmentCategory{
		ID:    row.ID,
		OrgID: row.OrgID,
		Name:  row.Name,
	}
}
