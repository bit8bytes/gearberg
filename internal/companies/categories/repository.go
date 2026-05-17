package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	genec "github.com/bit8bytes/gearberg/database/queries/gen/equipmentcategories"
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

// GetByCompanyID returns all equipmentCategories belonging to companyID.
func (r *Repository) GetByCompanyID(ctx context.Context, companyID string) ([]EquipmentCategory, error) {
	rows, err := r.equipmentCategories.GetEquipmentCategoriesByCompanyID(ctx, sql.NullString{String: companyID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("GetByCompanyID: %w", err)
	}
	result := make([]EquipmentCategory, len(rows))
	for i, row := range rows {
		result[i] = toModel(row)
	}
	return result, nil
}

// GetByID returns the category with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.GetEquipmentCategoryByID(ctx, id)
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
	row, err := r.equipmentCategories.CreateEquipmentCategory(ctx, genec.CreateEquipmentCategoryParams{
		ID:        c.ID,
		CompanyID: sql.NullString{String: c.CompanyID, Valid: true},
		Name:      c.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("Create: %w", database.NormalizeError(err))
	}
	m := EquipmentCategory{
		ID:        row.ID,
		CompanyID: database.String(row.CompanyID),
		Name:      row.Name,
	}
	return &m, nil
}

// Update updates the name of the category identified by u.ID.
func (r *Repository) Update(ctx context.Context, u UpdateEquipmentCategory) (*EquipmentCategory, error) {
	row, err := r.equipmentCategories.UpdateEquipmentCategory(ctx, genec.UpdateEquipmentCategoryParams{
		ID:   u.ID,
		Name: u.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("Update: %w", err)
	}
	m := toModel(row)
	return &m, nil
}

// Delete removes the category. Returns database.ErrForeignKeyViolation when inventory
// items are still assigned to the category.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.equipmentCategories.DeleteEquipmentCategory(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

func toModel(row genec.EquipmentCategory) EquipmentCategory {
	return EquipmentCategory{
		ID:        row.ID,
		CompanyID: database.String(row.CompanyID),
		Name:      row.Name,
	}
}
