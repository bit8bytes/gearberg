package inventory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genbs "github.com/bit8bytes/gearberg/internal/database/queries/gen/bulkstock"
	geninv "github.com/bit8bytes/gearberg/internal/database/queries/gen/inventory"
	gensu "github.com/bit8bytes/gearberg/internal/database/queries/gen/serializedunits"
	genui "github.com/bit8bytes/gearberg/internal/database/queries/gen/unitinspections"
	"github.com/bit8bytes/gearberg/internal/inventory/units"
	"github.com/bit8bytes/gearberg/internal/pagination"
)

// Repository provides data access for inventory items.
type Repository struct {
	db              *sql.DB
	inventory       *geninv.Queries
	serializedUnits *gensu.Queries
	bulkStock       *genbs.Queries
	unitInspections *genui.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:              db,
		inventory:       geninv.New(db),
		serializedUnits: gensu.New(db),
		bulkStock:       genbs.New(db),
		unitInspections: genui.New(db),
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
			Type:            Type(row.TypeID),
			UsageType:       UsageType(row.UsageTypeID),
			Name:            row.Name,
			Code:            row.Code,
			CategoryID:      row.CategoryID,
			CategoryName:    row.CategoryName,
			CategoryColor:   row.CategoryColor,
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
		ID:                     row.ID,
		OrgID:                  row.OrgID,
		Type:                   Type(row.TypeID),
		UsageType:              UsageType(row.UsageTypeID),
		Name:                   row.Name,
		Code:                   row.Code,
		CategoryID:             row.CategoryID,
		ManufacturerID:         database.String(row.ManufacturerID),
		StorageObjectID:        database.StringPtr(row.StorageObjectID),
		TotalStock:             row.TotalStock,
		PurchasePrice:          database.Int64Ptr(row.PurchasePrice),
		RentalPrice:            database.Int64Ptr(row.RentalPrice),
		Notes:                  database.String(row.Notes),
		WeightG:                database.Int64Ptr(row.WeightG),
		WidthMM:                database.Int64Ptr(row.WidthMm),
		HeightMM:               database.Int64Ptr(row.HeightMm),
		DepthMM:                database.Int64Ptr(row.DepthMm),
		PowerMW:                database.Int64Ptr(row.PowerMw),
		CurrentMA:              database.Int64Ptr(row.CurrentMa),
		InspectionIntervalDays: database.Int64Ptr(row.InspectionIntervalDays),
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
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

// createBulkWith inserts an inventory row and its bulk_stock row using the provided queries.
// Callers are responsible for transaction management.
func (r *Repository) createBulkWith(ctx context.Context, invQ *geninv.Queries, bsQ *genbs.Queries, c CreateBulkInventory) (*Inventory, error) {
	row, err := invQ.Create(ctx, geninv.CreateParams{
		ID:             c.ID,
		OrgID:          c.OrgID,
		Name:           c.Name,
		CategoryID:     c.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(c.ManufacturerID)),
		TypeID:         Bulk.ID(),
		UsageTypeID:    c.UsageTypeID,
		PurchasePrice:  database.NullInt64Ptr(c.PurchasePrice),
		RentalPrice:    database.NullInt64Ptr(c.RentalPrice),
		Notes:          database.NullString(database.StringOrNil(c.Notes)),
		WeightG:        database.NullInt64Ptr(c.WeightG),
		WidthMm:        database.NullInt64Ptr(c.WidthMM),
		HeightMm:       database.NullInt64Ptr(c.HeightMM),
		DepthMm:        database.NullInt64Ptr(c.DepthMM),
		PowerMw:        database.NullInt64Ptr(c.PowerMW),
		CurrentMa:      database.NullInt64Ptr(c.CurrentMA),
	})
	if err != nil {
		return nil, fmt.Errorf("createBulkWith: %w", database.NormalizeError(err))
	}
	if err := bsQ.Create(ctx, genbs.CreateParams{
		InventoryID: row.ID,
		Quantity:    c.TotalStock,
	}); err != nil {
		return nil, fmt.Errorf("createBulkWith: bulk_stock: %w", database.NormalizeError(err))
	}
	m := Inventory{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            Type(row.TypeID),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		Code:            row.Code,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      c.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		WeightG:         database.Int64Ptr(row.WeightG),
		WidthMM:         database.Int64Ptr(row.WidthMm),
		HeightMM:        database.Int64Ptr(row.HeightMm),
		DepthMM:         database.Int64Ptr(row.DepthMm),
		PowerMW:         database.Int64Ptr(row.PowerMw),
		CurrentMA:       database.Int64Ptr(row.CurrentMa),
		CreatedAt:       row.CreatedAt,
	}
	return &m, nil
}

// CreateBulk inserts a new bulk inventory item and its bulk_stock row atomically.
func (r *Repository) CreateBulk(ctx context.Context, c CreateBulkInventory) (*Inventory, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	m, err := r.createBulkWith(ctx, r.inventory.WithTx(tx), r.bulkStock.WithTx(tx), c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("CreateBulk: commit: %w", err)
	}
	return m, nil
}

// CreateBulkTx inserts a new bulk inventory item within an existing transaction.
func (r *Repository) CreateBulkTx(ctx context.Context, tx *sql.Tx, c CreateBulkInventory) (*Inventory, error) {
	m, err := r.createBulkWith(ctx, r.inventory.WithTx(tx), r.bulkStock.WithTx(tx), c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulkTx: %w", err)
	}
	return m, nil
}

// CreateSerialized inserts a serialized inventory item and all its units atomically within tx.
func (r *Repository) CreateSerialized(ctx context.Context, tx *sql.Tx, c CreateSerializedInventory) (*Inventory, error) {
	invQ := r.inventory.WithTx(tx)
	suQ := r.serializedUnits.WithTx(tx)
	row, err := invQ.Create(ctx, geninv.CreateParams{
		ID:             c.ID,
		OrgID:          c.OrgID,
		Name:           c.Name,
		CategoryID:     c.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(c.ManufacturerID)),
		TypeID:         Serialized.ID(),
		UsageTypeID:    c.UsageTypeID,
		PurchasePrice:  database.NullInt64Ptr(c.PurchasePrice),
		RentalPrice:    database.NullInt64Ptr(c.RentalPrice),
		Notes:          database.NullString(database.StringOrNil(c.Notes)),
		WeightG:        database.NullInt64Ptr(c.WeightG),
		WidthMm:        database.NullInt64Ptr(c.WidthMM),
		HeightMm:       database.NullInt64Ptr(c.HeightMM),
		DepthMm:        database.NullInt64Ptr(c.DepthMM),
		PowerMw:        database.NullInt64Ptr(c.PowerMW),
		CurrentMa:      database.NullInt64Ptr(c.CurrentMA),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateSerialized: %w", database.NormalizeError(err))
	}

	for _, u := range c.Units {
		_, err := suQ.CreateUnit(ctx, gensu.CreateUnitParams{
			ID:           u.ID,
			InventoryID:  row.ID,
			StatusID:     int64(units.UnitAvailable),
			SerialNumber: database.NullString(database.StringOrNil(u.SerialNumber)),
		})
		if err != nil {
			return nil, fmt.Errorf("CreateSerialized: create unit: %w", database.NormalizeError(err))
		}
	}

	m := Inventory{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            Type(row.TypeID),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		Code:            row.Code,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      int64(len(c.Units)),
		PurchasePrice:   database.Int64Ptr(row.PurchasePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		WeightG:         database.Int64Ptr(row.WeightG),
		WidthMM:         database.Int64Ptr(row.WidthMm),
		HeightMM:        database.Int64Ptr(row.HeightMm),
		DepthMM:         database.Int64Ptr(row.DepthMm),
		PowerMW:         database.Int64Ptr(row.PowerMw),
		CurrentMA:       database.Int64Ptr(row.CurrentMa),
		CreatedAt:       row.CreatedAt,
	}
	return &m, nil
}

// UpdateDetails updates the details-tab columns for a serialized inventory item.
func (r *Repository) UpdateDetails(ctx context.Context, u UpdateInventoryDetails) error {
	if err := r.inventory.UpdateDetails(ctx, geninv.UpdateDetailsParams{
		ID:             u.ID,
		Name:           u.Name,
		CategoryID:     u.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(u.ManufacturerID)),
		Notes:          database.NullString(database.StringOrNil(u.Notes)),
	}); err != nil {
		return fmt.Errorf("UpdateDetails: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateDetailsBulk updates the details-tab columns and bulk_stock quantity atomically.
func (r *Repository) UpdateDetailsBulk(ctx context.Context, u UpdateInventoryDetails) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateDetailsBulk: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if err := r.inventory.WithTx(tx).UpdateDetails(ctx, geninv.UpdateDetailsParams{
		ID:             u.ID,
		Name:           u.Name,
		CategoryID:     u.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(u.ManufacturerID)),
		Notes:          database.NullString(database.StringOrNil(u.Notes)),
	}); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: %w", database.NormalizeError(err))
	}
	if err := r.bulkStock.WithTx(tx).SetQuantity(ctx, genbs.SetQuantityParams{
		InventoryID: u.ID,
		Quantity:    u.TotalStock,
	}); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: bulk_stock: %w", database.NormalizeError(err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: commit: %w", err)
	}
	return nil
}

// UpdatePricing updates the pricing-tab columns for an inventory item.
func (r *Repository) UpdatePricing(ctx context.Context, u UpdateInventoryPricing) error {
	if err := r.inventory.UpdatePricing(ctx, geninv.UpdatePricingParams{
		ID:            u.ID,
		PurchasePrice: database.NullInt64Ptr(u.PurchasePrice),
		RentalPrice:   database.NullInt64Ptr(u.RentalPrice),
	}); err != nil {
		return fmt.Errorf("UpdatePricing: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateProperties updates the properties-tab columns for an inventory item.
func (r *Repository) UpdateProperties(ctx context.Context, u UpdateInventoryProperties) error {
	if err := r.inventory.UpdateProperties(ctx, geninv.UpdatePropertiesParams{
		ID:        u.ID,
		WeightG:   database.NullInt64Ptr(u.WeightG),
		WidthMm:   database.NullInt64Ptr(u.WidthMM),
		HeightMm:  database.NullInt64Ptr(u.HeightMM),
		DepthMm:   database.NullInt64Ptr(u.DepthMM),
		PowerMw:   database.NullInt64Ptr(u.PowerMW),
		CurrentMa: database.NullInt64Ptr(u.CurrentMA),
	}); err != nil {
		return fmt.Errorf("UpdateProperties: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateInspection updates the inspection_interval_days column for an inventory item.
func (r *Repository) UpdateInspection(ctx context.Context, u UpdateInventoryInspection) error {
	if err := r.inventory.UpdateInspection(ctx, geninv.UpdateInspectionParams{
		ID:                     u.ID,
		InspectionIntervalDays: database.NullInt64Ptr(u.InspectionIntervalDays),
	}); err != nil {
		return fmt.Errorf("UpdateInspection: %w", database.NormalizeError(err))
	}
	return nil
}

// ListUnits returns all serialized units for the given inventory item, ordered by unit_number.
func (r *Repository) ListUnits(ctx context.Context, inventoryID string) ([]Unit, error) {
	rows, err := r.serializedUnits.ListByInventoryID(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}
	us := make([]Unit, 0, len(rows))
	for _, row := range rows {
		us = append(us, Unit{
			ID:           row.ID,
			InventoryID:  row.InventoryID,
			StatusID:     row.StatusID,
			UnitNumber:   row.UnitNumber,
			SerialNumber: database.String(row.SerialNumber),
			Notes:        database.String(row.Notes),
			PurchasedAt:  database.Int64Ptr(row.PurchasedAt),
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return us, nil
}

// ListAll returns all inventory items for orgID ordered by name, with no pagination.
func (r *Repository) ListAll(ctx context.Context, orgID string) ([]Inventory, error) {
	rows, err := r.inventory.ListAllByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListAll: %w", err)
	}
	items := make([]Inventory, 0, len(rows))
	for _, row := range rows {
		items = append(items, Inventory{
			ID:              row.ID,
			OrgID:           row.OrgID,
			Type:            Type(row.TypeID),
			UsageType:       UsageType(row.UsageTypeID),
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
	return items, nil
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
	row, err := r.serializedUnits.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	u := Unit{
		ID:           row.ID,
		InventoryID:  row.InventoryID,
		StatusID:     row.StatusID,
		UnitNumber:   row.UnitNumber,
		SerialNumber: database.String(row.SerialNumber),
		Notes:        database.String(row.Notes),
		PurchasedAt:  database.Int64Ptr(row.PurchasedAt),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	return &u, nil
}

// AddUnit inserts a new empty unit for the inventory item.
func (r *Repository) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
	row, err := r.serializedUnits.CreateUnit(ctx, gensu.CreateUnitParams{
		ID:           a.ID,
		InventoryID:  a.InventoryID,
		StatusID:     int64(units.UnitAvailable),
		SerialNumber: database.NullString(nil),
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

// ListUnitStatuses returns all rows from the unit_statuses lookup table.
func (r *Repository) ListUnitStatuses(ctx context.Context) ([]UnitStatusEntry, error) {
	rows, err := r.inventory.ListUnitStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListUnitStatuses: %w", err)
	}
	out := make([]UnitStatusEntry, len(rows))
	for i, row := range rows {
		out[i] = UnitStatusEntry{ID: row.ID, Name: row.Name}
	}
	return out, nil
}

// UpdateUnit updates the editable fields of a unit.
func (r *Repository) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := r.serializedUnits.Update(ctx, gensu.UpdateParams{
		StatusID:     u.StatusID,
		SerialNumber: database.NullString(database.StringOrNil(u.SerialNumber)),
		Notes:        database.NullString(database.StringOrNil(u.Notes)),
		PurchasedAt:  database.NullInt64Ptr(u.PurchasedAt),
		ID:           u.ID,
	}); err != nil {
		return fmt.Errorf("UpdateUnit: %w", database.NormalizeError(err))
	}
	return nil
}

// DeleteUnit removes a unit by ID.
func (r *Repository) DeleteUnit(ctx context.Context, id string) error {
	if err := r.serializedUnits.Delete(ctx, id); err != nil {
		return fmt.Errorf("DeleteUnit: %w", database.NormalizeError(err))
	}
	return nil
}

// LogInspection inserts a new inspection entry for a unit.
func (r *Repository) LogInspection(ctx context.Context, l LogInspection) (*Inspection, error) {
	var passed int64
	if l.Passed {
		passed = 1
	}
	row, err := r.unitInspections.CreateInspection(ctx, genui.CreateInspectionParams{
		ID:          l.ID,
		UnitID:      l.UnitID,
		InspectedAt: l.InspectedAt,
		Passed:      passed,
		Notes:       database.NullString(database.StringOrNil(l.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("LogInspection: %w", database.NormalizeError(err))
	}
	return &Inspection{
		ID:          row.ID,
		UnitID:      row.UnitID,
		InspectedAt: row.InspectedAt,
		Passed:      row.Passed == 1,
		Notes:       database.String(row.Notes),
		CreatedAt:   row.CreatedAt,
	}, nil
}

// ListLatestInspectionAtByInventoryID returns a map of unit ID → most recent
// inspection timestamp for all units belonging to inventoryID.
func (r *Repository) ListLatestInspectionAtByInventoryID(ctx context.Context, inventoryID string) (map[string]int64, error) {
	rows, err := r.unitInspections.ListLatestInspectedAtByInventoryID(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("ListLatestInspectionAtByInventoryID: %w", err)
	}
	m := make(map[string]int64, len(rows))
	for _, row := range rows {
		at, ok := row.InspectedAt.(int64)
		if !ok {
			continue
		}
		m[row.UnitID] = at
	}
	return m, nil
}

// ListInspections returns all inspection entries for a unit, newest first.
func (r *Repository) ListInspections(ctx context.Context, unitID string) ([]Inspection, error) {
	rows, err := r.unitInspections.ListByUnitID(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("ListInspections: %w", err)
	}
	out := make([]Inspection, len(rows))
	for i, row := range rows {
		out[i] = Inspection{
			ID:          row.ID,
			UnitID:      row.UnitID,
			InspectedAt: row.InspectedAt,
			Passed:      row.Passed == 1,
			Notes:       database.String(row.Notes),
			CreatedAt:   row.CreatedAt,
		}
	}
	return out, nil
}
