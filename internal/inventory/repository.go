package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	geninv "github.com/bit8bytes/gearberg/database/queries/gen/inventory"
	"github.com/bit8bytes/gearberg/internal/inventory/types"
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

// MaxCode returns the highest code for orgID, or 0 when there are no items.
func (r *Repository) MaxCode(ctx context.Context, orgID string) (int64, error) {
	n, err := r.inventory.MaxCodeByOrgID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("MaxCode: %w", err)
	}
	return n, nil
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
			TypeID:          row.TypeID,
			UsageTypeID:     row.UsageTypeID,
			Name:            row.Name,
			Code:            row.Code,
			CategoryID:      row.CategoryID,
			CategoryName:    row.CategoryName,
			ManufacturerID:  database.String(row.ManufacturerID),
			StorageObjectID: database.StringPtr(row.StorageObjectID),
			TotalStock:      row.TotalStock,
			PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
			RentalPrice:     database.Int64Ptr(row.RentalPrice),
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
		TypeID:          row.TypeID,
		UsageTypeID:     row.UsageTypeID,
		Name:            row.Name,
		Code:            row.Code,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
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

// CreateBulk inserts a new bulk inventory item.
func (r *Repository) CreateBulk(ctx context.Context, c CreateBulkInventory) (*Inventory, error) {
	row, err := r.inventory.Create(ctx, geninv.CreateParams{
		ID:            c.ID,
		OrgID:         c.OrgID,
		Name:          c.Name,
		CategoryID:    c.CategoryID,
		TypeID:        types.Bulk.ID(),
		UsageTypeID:   c.UsageTypeID,
		Code:          c.Code,
		TotalStock:    c.TotalStock,
		PurchasePrice: database.NullInt64Ptr(c.PurchasePrice),
		RentalPrice:   database.NullInt64Ptr(c.RentalPrice),
		Notes:         database.NullString(database.StringOrNil(c.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", database.NormalizeError(err))
	}
	m := Inventory{
		ID:              row.ID,
		OrgID:           row.OrgID,
		TypeID:          row.TypeID,
		UsageTypeID:     row.UsageTypeID,
		Name:            row.Name,
		Code:            row.Code,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		CreatedAt:       row.CreatedAt,
	}
	return &m, nil
}

// CreateSerialized inserts a serialized inventory item and all its units atomically within tx.
func (r *Repository) CreateSerialized(ctx context.Context, tx *sql.Tx, c CreateSerializedInventory) (*Inventory, error) {
	q := geninv.New(tx)
	row, err := q.Create(ctx, geninv.CreateParams{
		ID:            c.ID,
		OrgID:         c.OrgID,
		Name:          c.Name,
		CategoryID:    c.CategoryID,
		TypeID:        types.Serialized.ID(),
		UsageTypeID:   c.UsageTypeID,
		Code:          c.Code,
		TotalStock:    int64(len(c.Units)),
		PurchasePrice: database.NullInt64Ptr(c.PurchasePrice),
		RentalPrice:   database.NullInt64Ptr(c.RentalPrice),
		Notes:         database.NullString(database.StringOrNil(c.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateSerialized: %w", database.NormalizeError(err))
	}

	for _, u := range c.Units {
		_, err := q.CreateUnit(ctx, geninv.CreateUnitParams{
			ID:               u.ID,
			InventoryID:      row.ID,
			StatusID:         int64(types.UnitAvailable),
			UnitNumber:       u.UnitNumber,
			SerialNumber:     database.NullString(database.StringOrNil(u.SerialNumber)),
			NextInspectionAt: database.NullInt64(u.NextInspectionAt),
		})
		if err != nil {
			return nil, fmt.Errorf("CreateSerialized: create unit %d: %w", u.UnitNumber, database.NormalizeError(err))
		}
	}

	m := Inventory{
		ID:              row.ID,
		OrgID:           row.OrgID,
		TypeID:          row.TypeID,
		UsageTypeID:     row.UsageTypeID,
		Name:            row.Name,
		Code:            row.Code,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
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
		Code:          u.Code,
		TotalStock:    u.TotalStock,
		PurchasePrice: database.NullInt64Ptr(u.PurchasePrice),
		RentalPrice:   database.NullInt64Ptr(u.RentalPrice),
		Notes:         database.NullString(database.StringOrNil(u.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("Update: %w", database.NormalizeError(err))
	}
	m := Inventory{
		ID:              row.ID,
		OrgID:           row.OrgID,
		TypeID:          row.TypeID,
		UsageTypeID:     row.UsageTypeID,
		Name:            row.Name,
		Code:            row.Code,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	return &m, nil
}

// ListUnits returns all inventory units for the given inventory item, ordered by unit_number.
func (r *Repository) ListUnits(ctx context.Context, inventoryID string) ([]Unit, error) {
	rows, err := r.inventory.ListUnitsByInventoryID(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}
	units := make([]Unit, 0, len(rows))
	for _, row := range rows {
		units = append(units, Unit{
			ID:               row.ID,
			InventoryID:      row.InventoryID,
			StatusID:         row.StatusID,
			UnitNumber:       row.UnitNumber,
			SerialNumber:     database.String(row.SerialNumber),
			Notes:            database.String(row.Notes),
			NextInspectionAt: database.Int64Ptr(row.NextInspectionAt),
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		})
	}
	return units, nil
}

// Delete removes the inventory item. Returns database.ErrForeignKeyViolation when
// active rental line items reference it.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.inventory.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

// GetUnit returns the unit with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetUnit(ctx context.Context, id string) (*Unit, error) {
	row, err := r.inventory.GetUnit(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	u := Unit{
		ID:               row.ID,
		InventoryID:      row.InventoryID,
		StatusID:         row.StatusID,
		UnitNumber:       row.UnitNumber,
		SerialNumber:     database.String(row.SerialNumber),
		Notes:            database.String(row.Notes),
		NextInspectionAt: database.Int64Ptr(row.NextInspectionAt),
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	return &u, nil
}

// MaxUnitNumber returns the highest unit_number for inventoryID, or 0 when there are no units.
func (r *Repository) MaxUnitNumber(ctx context.Context, inventoryID string) (int64, error) {
	n, err := r.inventory.MaxUnitNumber(ctx, inventoryID)
	if err != nil {
		return 0, fmt.Errorf("MaxUnitNumber: %w", err)
	}
	return n, nil
}

// AddUnit inserts a new empty unit for the inventory item.
func (r *Repository) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
	row, err := r.inventory.CreateUnit(ctx, geninv.CreateUnitParams{
		ID:               a.ID,
		InventoryID:      a.InventoryID,
		StatusID:         int64(types.UnitAvailable),
		UnitNumber:       a.UnitNumber,
		SerialNumber:     database.NullString(nil),
		NextInspectionAt: database.NullInt64(nil),
	})
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", database.NormalizeError(err))
	}
	u := Unit{
		ID:          row.ID,
		InventoryID: row.InventoryID,
		StatusID:    row.StatusID,
		UnitNumber:  row.UnitNumber,
		CreatedAt:   row.CreatedAt,
	}
	return &u, nil
}

// UpdateUnit updates the serial number and next inspection date of a unit.
func (r *Repository) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := r.inventory.UpdateUnit(ctx, geninv.UpdateUnitParams{
		SerialNumber:     database.NullString(database.StringOrNil(u.SerialNumber)),
		NextInspectionAt: database.NullInt64(u.NextInspectionAt),
		Notes:            database.NullString(database.StringOrNil(u.Notes)),
		ID:               u.ID,
	}); err != nil {
		return fmt.Errorf("UpdateUnit: %w", database.NormalizeError(err))
	}
	return nil
}

// DeleteUnit removes a unit by ID.
func (r *Repository) DeleteUnit(ctx context.Context, id string) error {
	if err := r.inventory.DeleteUnit(ctx, id); err != nil {
		return fmt.Errorf("DeleteUnit: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateTotalStock increments or decrements total_stock for the inventory item by delta.
func (r *Repository) UpdateTotalStock(ctx context.Context, id string, delta int64) error {
	if err := r.inventory.UpdateTotalStock(ctx, geninv.UpdateTotalStockParams{
		TotalStock: delta,
		ID:         id,
	}); err != nil {
		return fmt.Errorf("UpdateTotalStock: %w", err)
	}
	return nil
}
