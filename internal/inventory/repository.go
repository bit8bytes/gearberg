package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	geninv "github.com/bit8bytes/gearberg/database/queries/gen/inventory"
	"github.com/bit8bytes/gearberg/internal/pagination"
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

// Count returns the number of inventory items belonging to orgID.
func (r *Repository) Count(ctx context.Context, orgID string) (int64, error) {
	n, err := r.inventory.CountByOrgID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// List returns a page of inventory items for orgID. When query is non-empty, only
// items whose name contains query (case-insensitive) are returned. When category is
// non-empty, only items in that category are returned. Returns total matching count.
func (r *Repository) List(ctx context.Context, orgID, query, category string, f pagination.Filters) ([]Inventory, int, error) {
	rows, err := r.inventory.List(ctx, geninv.ListParams{
		OrgID:      orgID,
		NameQuery:  query,
		Category:   category,
		PageOffset: int64(f.Offset()),
		PageLimit:  int64(f.Limit()),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("List: %w", err)
	}

	var totalRecords int64
	items := make([]Inventory, 0, len(rows))
	for _, row := range rows {
		totalRecords = row.TotalRecords
		items = append(items, Inventory{
			ID:              row.ID,
			OrgID:           row.OrgID,
			Name:            row.Name,
			CategoryID:      row.CategoryID,
			CategoryName:    row.CategoryName,
			ManufacturerID:  database.String(row.ManufacturerID),
			StorageObjectID: database.StringPtr(row.StorageObjectID),
			TotalStock:      row.TotalStock,
			PurchasePrice:   database.Float64Ptr(row.PurchasePrice),
			RentalPrice:     database.Float64Ptr(row.RentalPrice),
			Notes:           database.String(row.Notes),
			UpdatedAt:       row.UpdatedAt,
			CreatedAt:       row.CreatedAt,
		})
	}
	return items, int(totalRecords), nil
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
		OrgID:           row.OrgID,
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
		OrgID:         c.OrgID,
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
		OrgID:           row.OrgID,
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
		OrgID:           row.OrgID,
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
