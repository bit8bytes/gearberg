package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	geninv "github.com/bit8bytes/gearberg/database/queries/gen/inventory"
)

// Repository provides data access for inventory items.
type Repository struct {
	db        *sql.DB
	inventory geninv.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:        db,
		inventory: geninv.New(db),
	}
}

// Count returns the number of inventory items belonging to companyID.
func (r *Repository) Count(ctx context.Context, companyID string) (int64, error) {
	n, err := r.inventory.CountByCompanyID(ctx, companyID)
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// GetByCompanyID returns all inventory items belonging to companyID.
func (r *Repository) GetByCompanyID(ctx context.Context, companyID string) ([]Inventory, error) {
	rows, err := r.inventory.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("GetByCompanyID: %w", err)
	}
	result := make([]Inventory, len(rows))
	for i, row := range rows {
		result[i] = toListModel(row)
	}
	return result, nil
}

// GetByID returns the inventory item with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*Inventory, error) {
	row, err := r.inventory.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := Inventory{
		ID:              row.ID,
		CompanyID:       row.CompanyID,
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Float64Ptr(row.PurchasePrice),
		RentalPrice:     database.Float64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	return &m, nil
}

// SetImage links or unlinks a storage object from an inventory item.
func (r *Repository) SetImage(ctx context.Context, s SetImage) error {
	if err := r.inventory.UpdateStorageObject(ctx, geninv.UpdateStorageObjectParams{
		ID:              s.ID,
		StorageObjectID: database.NullString(s.StorageObjectID),
	}); err != nil {
		return fmt.Errorf("SetImage: %w", err)
	}
	return nil
}

// Create inserts a new inventory item.
func (r *Repository) Create(ctx context.Context, c CreateInventory) (*Inventory, error) {
	row, err := r.inventory.Create(ctx, geninv.CreateParams{
		ID:            c.ID,
		CompanyID:     c.CompanyID,
		Name:          c.Name,
		CategoryID:    c.CategoryID,
		TotalStock:    c.TotalStock,
		PurchasePrice: database.NullFloat64Ptr(c.PurchasePrice),
		RentalPrice:   database.NullFloat64Ptr(c.RentalPrice),
		Notes:         database.NullString(database.StringOrNil(c.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("Create: %w", database.NormalizeError(err))
	}
	m := Inventory{
		ID:              row.ID,
		CompanyID:       row.CompanyID,
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Float64Ptr(row.PurchasePrice),
		RentalPrice:     database.Float64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		CreatedAt:       row.CreatedAt,
	}
	return &m, nil
}

// Update updates the fields of the inventory item identified by u.ID.
func (r *Repository) Update(ctx context.Context, u UpdateInventory) (*Inventory, error) {
	row, err := r.inventory.Update(ctx, geninv.UpdateParams{
		ID:            u.ID,
		Name:          u.Name,
		CategoryID:    u.CategoryID,
		TotalStock:    u.TotalStock,
		PurchasePrice: database.NullFloat64Ptr(u.PurchasePrice),
		RentalPrice:   database.NullFloat64Ptr(u.RentalPrice),
		Notes:         database.NullString(database.StringOrNil(u.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("Update: %w", database.NormalizeError(err))
	}
	m := Inventory{
		ID:              row.ID,
		CompanyID:       row.CompanyID,
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Float64Ptr(row.PurchasePrice),
		RentalPrice:     database.Float64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	return &m, nil
}

// Delete removes the inventory item. Returns database.ErrForeignKeyViolation when
// active rental line items reference it.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.inventory.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

func toListModel(row geninv.GetByCompanyIDRow) Inventory {
	return Inventory{
		ID:              row.ID,
		CompanyID:       row.CompanyID,
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		CategoryName:    row.CategoryName,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Float64Ptr(row.PurchasePrice),
		RentalPrice:     database.Float64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
