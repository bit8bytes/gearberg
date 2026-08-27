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

// Package equipment provides equipment functionality.
package equipment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genequip "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipment"
	genbulk "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentbulkitems"
	gencontent "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentcombinationitems"
	genserialized "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentserializeditems"
	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
	"github.com/bit8bytes/gearberg/internal/equipment/usage"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/units"
)

// Repository provides data access for inventory items.
type Repository struct {
	db              *sql.DB
	equipment       *genequip.Queries
	serializedItems *genserialized.Queries
	bulkItems       *genbulk.Queries
	content         *gencontent.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:              db,
		equipment:       genequip.New(db),
		serializedItems: genserialized.New(db),
		bulkItems:       genbulk.New(db),
		content:         gencontent.New(db),
	}
}

// Count returns the number of inventory items belonging to orgID.
func (r *Repository) Count(ctx context.Context, orgID string) (int64, error) {
	n, err := r.equipment.CountByOrgID(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("Count: %w", err)
	}
	return n, nil
}

// List returns a page of inventory items for orgID. When query is non-empty, only
// items whose name or unit code contains query are returned. When category is
// non-empty, only items in that category are returned. When showArchived is true,
// only archived items are returned; otherwise only active items are returned.
// sortBy may be "code" to order by minimum unit code; otherwise orders by name.
// Returns total matching count.
func (r *Repository) List(ctx context.Context, orgID, query, category, inspectionFilter string, showArchived bool, f pagination.Filters) ([]Equipment, int, error) {
	rows, err := r.equipment.List(ctx, genequip.ListParams{
		OrgID:            orgID,
		NameQuery:        query,
		Category:         category,
		InspectionFilter: inspectionFilter,
		IsArchived:       database.Bool(showArchived),
		PageOffset:       int64(f.Offset()),
		PageLimit:        int64(f.Limit()),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("List: %w", err)
	}
	return listRowsToEquipment(rows)
}

func listRowsToEquipment(rows []genequip.ListRow) ([]Equipment, int, error) {
	var totalRecords int64
	items := make([]Equipment, 0, len(rows))
	for _, row := range rows {
		totalRecords = row.TotalRecords
		items = append(items, Equipment{
			ID:              row.ID,
			OrgID:           row.OrgID,
			Type:            Parse(row.EquipmentTypeName),
			TrackingType:    tracking.Type(row.TrackingTypeID.Int64),
			UsageType:       usage.Type(row.UsageTypeID),
			Name:            row.Name,
			CategoryID:      database.String(row.CategoryID),
			CategoryName:    row.CategoryName,
			ManufacturerID:  database.String(row.ManufacturerID),
			LocationID:      database.String(row.LocationID),
			LocationName:    row.LocationName,
			StorageObjectID: database.StringPtr(row.StorageObjectID),
			TotalStock:      row.TotalStock,
			IsArchived:      row.IsArchived == 1,
			Notes:           database.String(row.Notes),
			Pricing: Pricing{
				PurchasePrice: row.ResalePrice,
				RentalPrice:   row.RentalPrice,
			},
			Properties: Properties{
				Weight:    row.WeightG,
				Width:     row.WidthMm,
				Height:    row.HeightMm,
				Depth:     row.DepthMm,
				Power:     row.PowerMw,
				Current:   row.CurrentMa,
				Voltage:   row.VoltageMv,
				WireGauge: row.WireGaugeMm2X100,
			},
			UpdatedAt: row.UpdatedAt,
			CreatedAt: row.CreatedAt,
			InspectionStatus: func() InspectionStatus {
				v, ok := row.MinNextInspectionAt.(int64)
				if !ok {
					return InspectionStatusNone
				}
				return NewInspectionStatus(&v)
			}(),
		})
	}
	return items, int(totalRecords), nil
}

// Get returns the inventory item with id, or ErrNotFound when it does not exist.
func (r *Repository) Get(ctx context.Context, id string) (*Equipment, error) {
	row, err := r.equipment.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            Parse(row.EquipmentTypeName),
		TrackingType:    tracking.Type(row.TrackingTypeID.Int64),
		UsageType:       usage.Type(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      database.String(row.CategoryID),
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		LocationName:    row.LocationName,
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		ContentCount:    row.ContentCount,
		IsArchived:      row.IsArchived == 1,
		Notes:           database.String(row.Notes),
		Pricing: Pricing{
			PurchasePrice: row.ResalePrice,
			RentalPrice:   row.RentalPrice,
		},
		Properties: Properties{
			Weight:    row.WeightG,
			Width:     row.WidthMm,
			Height:    row.HeightMm,
			Depth:     row.DepthMm,
			Power:     row.PowerMw,
			Current:   row.CurrentMa,
			Voltage:   row.VoltageMv,
			WireGauge: row.WireGaugeMm2X100,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if tracking.Type(row.TrackingTypeID.Int64) == tracking.Bulk {
		if bulk, err := r.bulkItems.GetByEquipmentID(ctx, row.ID); err == nil {
			m.BulkPurchasePrice = database.NullAs[units.Cents](bulk.PurchasePrice)
		}
	}
	return &m, nil
}

// SetImage links or unlinks a storage object from an inventory item.
func (r *Repository) SetImage(ctx context.Context, s SetImage) error {
	if err := r.equipment.UpdateStorageObject(ctx, genequip.UpdateStorageObjectParams{
		ID:              s.ID,
		StorageObjectID: database.NullString(s.StorageObjectID),
	}); err != nil {
		return fmt.Errorf("SetImage: %w", err)
	}
	return nil
}

// createBulkWith inserts an equipment row and its equipment_bulk_items row using the provided queries.
// Callers are responsible for transaction management.
func (r *Repository) createBulkWith(ctx context.Context, eqQ *genequip.Queries, bulkQ *genbulk.Queries, c CreateBulkEquipment) (*Equipment, error) {
	row, err := eqQ.Create(ctx, genequip.CreateParams{
		ID:               c.ID,
		OrgID:            c.OrgID,
		EquipmentTypeID:  c.EquipmentType.ID(),
		TrackingTypeID:   sql.NullInt64{Int64: tracking.Bulk.ID(), Valid: true},
		CategoryID:       database.NullString(database.StringOrNil(c.CategoryID)),
		ManufacturerID:   database.NullString(database.StringOrNil(c.ManufacturerID)),
		UsageTypeID:      c.UsageTypeID,
		LocationID:       database.NullString(database.StringOrNil(c.LocationID)),
		Name:             c.Name,
		IsArchived:       0,
		RentalPrice:      c.Pricing.RentalPrice,
		ResalePrice:      c.Pricing.PurchasePrice,
		Notes:            database.NullString(database.StringOrNil(c.Notes)),
		WeightG:          c.Properties.Weight,
		WidthMm:          c.Properties.Width,
		HeightMm:         c.Properties.Height,
		DepthMm:          c.Properties.Depth,
		CurrentMa:        c.Properties.Current,
		PowerMw:          c.Properties.Power,
		VoltageMv:        c.Properties.Voltage,
		WireGaugeMm2X100: c.Properties.WireGauge,
	})
	if err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("createBulkWith: %w", normalized)
	}
	if _, err := bulkQ.Create(ctx, genbulk.CreateParams{
		ID:            c.BulkItemID,
		EquipmentID:   row.ID,
		Quantity:      c.TotalStock,
		PurchasePrice: database.NullOf(c.Pricing.PurchasePrice),
	}); err != nil {
		return nil, fmt.Errorf("createBulkWith: equipment_bulk_items: %w", database.NormalizeError(err))
	}
	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            c.EquipmentType,
		TrackingType:    tracking.Type(row.TrackingTypeID.Int64),
		UsageType:       usage.Type(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      database.String(row.CategoryID),
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      c.TotalStock,
		Notes:           database.String(row.Notes),
		Pricing: Pricing{
			PurchasePrice: row.ResalePrice,
			RentalPrice:   row.RentalPrice,
		},
		Properties: Properties{
			Weight:    row.WeightG,
			Width:     row.WidthMm,
			Height:    row.HeightMm,
			Depth:     row.DepthMm,
			Power:     row.PowerMw,
			Current:   row.CurrentMa,
			Voltage:   row.VoltageMv,
			WireGauge: row.WireGaugeMm2X100,
		},
		CreatedAt: row.CreatedAt,
	}
	return &m, nil
}

// CreateBulk inserts a new bulk inventory item within an existing transaction.
func (r *Repository) CreateBulk(ctx context.Context, tx *sql.Tx, c CreateBulkEquipment) (*Equipment, error) {
	m, err := r.createBulkWith(ctx, r.equipment.WithTx(tx), r.bulkItems.WithTx(tx), c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulkTx: %w", err)
	}
	return m, nil
}

// CreateSerialized inserts a serialized inventory item and all its units atomically within tx.
func (r *Repository) CreateSerialized(ctx context.Context, tx *sql.Tx, c CreateSerializedEquipment) (*Equipment, error) {
	eqQ := r.equipment.WithTx(tx)
	itemQ := r.serializedItems.WithTx(tx)
	row, err := eqQ.Create(ctx, genequip.CreateParams{
		ID:               c.ID,
		OrgID:            c.OrgID,
		EquipmentTypeID:  c.EquipmentType.ID(),
		TrackingTypeID:   sql.NullInt64{Int64: tracking.Serialized.ID(), Valid: true},
		CategoryID:       database.NullString(database.StringOrNil(c.CategoryID)),
		ManufacturerID:   database.NullString(database.StringOrNil(c.ManufacturerID)),
		UsageTypeID:      c.UsageTypeID,
		LocationID:       database.NullString(database.StringOrNil(c.LocationID)),
		Name:             c.Name,
		IsArchived:       0,
		RentalPrice:      c.Pricing.RentalPrice,
		ResalePrice:      c.Pricing.PurchasePrice,
		Notes:            database.NullString(database.StringOrNil(c.Notes)),
		WeightG:          c.Properties.Weight,
		WidthMm:          c.Properties.Width,
		HeightMm:         c.Properties.Height,
		DepthMm:          c.Properties.Depth,
		CurrentMa:        c.Properties.Current,
		PowerMw:          c.Properties.Power,
		VoltageMv:        c.Properties.Voltage,
		WireGaugeMm2X100: c.Properties.WireGauge,
	})
	if err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("CreateSerialized: %w", normalized)
	}

	for _, u := range c.Units {
		if _, err := itemQ.Create(ctx, genserialized.CreateParams{
			ID:                 u.ID,
			OrgID:              c.OrgID,
			EquipmentID:        row.ID,
			SerialNumber:       u.SerialNumber,
			IsActive:           database.Bool(u.IsActive),
			Remark:             database.NullString(database.StringOrNil(u.Remark)),
			PurchasePrice:      database.NullOf(u.PurchasePrice),
			PurchasedAt:        database.NullInt64Ptr(u.PurchasedAt),
			NextInspectionAt:   database.NullInt64Ptr(u.NextInspectionAt),
			ManufacturerSerial: database.NullString(database.StringOrNil(u.ManufacturerSerialNumber)),
		}); err != nil {
			return nil, fmt.Errorf("CreateSerialized: create item: %w", database.NormalizeError(err))
		}
	}

	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            c.EquipmentType,
		TrackingType:    tracking.Type(row.TrackingTypeID.Int64),
		UsageType:       usage.Type(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      database.String(row.CategoryID),
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      int64(len(c.Units)),
		Notes:           database.String(row.Notes),
		Pricing: Pricing{
			PurchasePrice: row.ResalePrice,
			RentalPrice:   row.RentalPrice,
		},
		Properties: Properties{
			Weight:    row.WeightG,
			Width:     row.WidthMm,
			Height:    row.HeightMm,
			Depth:     row.DepthMm,
			Power:     row.PowerMw,
			Current:   row.CurrentMa,
			Voltage:   row.VoltageMv,
			WireGauge: row.WireGaugeMm2X100,
		},
		CreatedAt: row.CreatedAt,
	}
	return &m, nil
}

// UpdateDetails updates the details-tab columns for an inventory item.
func (r *Repository) UpdateDetails(ctx context.Context, u UpdateEquipmentDetails) error {
	if err := r.equipment.UpdateDetails(ctx, genequip.UpdateDetailsParams{
		ID:             u.ID,
		Name:           u.Name,
		CategoryID:     database.NullString(database.StringOrNil(u.CategoryID)),
		ManufacturerID: database.NullString(database.StringOrNil(u.ManufacturerID)),
		LocationID:     database.NullString(database.StringOrNil(u.LocationID)),
		Notes:          database.NullString(database.StringOrNil(u.Notes)),
	}); err != nil {
		return fmt.Errorf("UpdateDetails: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateDetailsBulk updates the details-tab columns and equipment_bulk_items quantity atomically.
func (r *Repository) UpdateDetailsBulk(ctx context.Context, u UpdateEquipmentDetails) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateDetailsBulk: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if err := r.equipment.WithTx(tx).UpdateDetails(ctx, genequip.UpdateDetailsParams{
		ID:             u.ID,
		Name:           u.Name,
		CategoryID:     database.NullString(database.StringOrNil(u.CategoryID)),
		ManufacturerID: database.NullString(database.StringOrNil(u.ManufacturerID)),
		LocationID:     database.NullString(database.StringOrNil(u.LocationID)),
		Notes:          database.NullString(database.StringOrNil(u.Notes)),
	}); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: %w", database.NormalizeError(err))
	}
	item, err := r.bulkItems.WithTx(tx).GetByEquipmentID(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("UpdateDetailsBulk: get bulk item: %w", err)
	}
	if err := r.bulkItems.WithTx(tx).SetQuantity(ctx, genbulk.SetQuantityParams{
		Quantity: u.TotalStock,
		ID:       item.ID,
	}); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: set quantity: %w", database.NormalizeError(err))
	}
	if err := r.bulkItems.WithTx(tx).SetPurchasePrice(ctx, genbulk.SetPurchasePriceParams{
		PurchasePrice: database.NullOf(u.PurchasePrice),
		ID:            item.ID,
	}); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: set purchase price: %w", database.NormalizeError(err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: commit: %w", err)
	}
	return nil
}

// UpdatePricing updates the pricing-tab columns for an inventory item.
func (r *Repository) UpdatePricing(ctx context.Context, u UpdateEquipmentPricing) error {
	if err := r.equipment.UpdatePricing(ctx, genequip.UpdatePricingParams{
		ID:          u.ID,
		ResalePrice: u.Pricing.PurchasePrice,
		RentalPrice: u.Pricing.RentalPrice,
	}); err != nil {
		return fmt.Errorf("UpdatePricing: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateProperties updates the properties-tab columns for an inventory item.
func (r *Repository) UpdateProperties(ctx context.Context, u UpdateEquipmentProperties) error {
	if err := r.equipment.UpdateProperties(ctx, genequip.UpdatePropertiesParams{
		ID:               u.ID,
		WeightG:          u.Properties.Weight,
		WidthMm:          u.Properties.Width,
		HeightMm:         u.Properties.Height,
		DepthMm:          u.Properties.Depth,
		VoltageMv:        u.Properties.Voltage,
		CurrentMa:        u.Properties.Current,
		PowerMw:          u.Properties.Power,
		WireGaugeMm2X100: u.Properties.WireGauge,
	}); err != nil {
		return fmt.Errorf("UpdateProperties: %w", database.NormalizeError(err))
	}
	return nil
}

// ListUnits returns all serialized units for the given inventory item, ordered by serial_number.
func (r *Repository) ListUnits(ctx context.Context, equipmentID string) ([]Unit, error) {
	rows, err := r.serializedItems.ListByEquipmentID(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}
	us := make([]Unit, 0, len(rows))
	for _, row := range rows {
		us = append(us, Unit{
			ID:                       row.ID,
			EquipmentID:              row.EquipmentID,
			StatusID:                 row.IsActive,
			SerialNumber:             row.SerialNumber,
			ManufacturerSerialNumber: database.String(row.ManufacturerSerial),
			Remark:                   database.String(row.Remark),
			PurchasePrice:            database.NullAs[units.Cents](row.PurchasePrice),
			PurchasedAt:              database.Int64Ptr(row.PurchasedAt),
			NextInspectionAt:         database.Int64Ptr(row.NextInspectionAt),
			CreatedAt:                row.CreatedAt,
			UpdatedAt:                row.UpdatedAt,
		})
	}
	return us, nil
}

// Delete removes the inventory item. Returns ErrInUse when active rental line items reference it.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.equipment.Delete(ctx, id); err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrForeignKeyViolation) {
			return ErrInUse
		}
		return fmt.Errorf("Delete: %w", normalized)
	}
	return nil
}

// GetUnit returns the serialized unit with id, or ErrNotFound when it does not exist.
func (r *Repository) GetUnit(ctx context.Context, id string) (*Unit, error) {
	row, err := r.serializedItems.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	u := Unit{
		ID:                       row.ID,
		EquipmentID:              row.EquipmentID,
		StatusID:                 row.IsActive,
		SerialNumber:             row.SerialNumber,
		ManufacturerSerialNumber: database.String(row.ManufacturerSerial),
		Remark:                   database.String(row.Remark),
		PurchasePrice:            database.NullAs[units.Cents](row.PurchasePrice),
		PurchasedAt:              database.Int64Ptr(row.PurchasedAt),
		NextInspectionAt:         database.Int64Ptr(row.NextInspectionAt),
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
	return &u, nil
}

// AddUnit inserts a new serialized unit for the inventory item.
func (r *Repository) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
	row, err := r.serializedItems.Create(ctx, genserialized.CreateParams{
		ID:           a.ID,
		OrgID:        a.OrgID,
		EquipmentID:  a.EquipmentID,
		SerialNumber: a.SerialNumber,
		IsActive:     1,
	})
	if err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("AddUnit: %w", normalized)
	}
	u := Unit{
		ID:           row.ID,
		EquipmentID:  row.EquipmentID,
		StatusID:     row.IsActive,
		SerialNumber: row.SerialNumber,
		CreatedAt:    row.CreatedAt,
	}
	return &u, nil
}

// UpdateUnit updates the editable fields of a serialized unit.
func (r *Repository) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := r.serializedItems.Update(ctx, genserialized.UpdateParams{
		SerialNumber:       u.SerialNumber,
		IsActive:           u.IsActive,
		Remark:             database.NullString(database.StringOrNil(u.Remark)),
		PurchasePrice:      database.NullOf(u.PurchasePrice),
		PurchasedAt:        database.NullInt64Ptr(u.PurchasedAt),
		NextInspectionAt:   database.NullInt64Ptr(u.NextInspectionAt),
		ManufacturerSerial: database.NullString(database.StringOrNil(u.ManufacturerSerialNumber)),
		ID:                 u.ID,
	}); err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return ErrConflict
		}
		return fmt.Errorf("UpdateUnit: %w", normalized)
	}
	return nil
}

// BulkUpdateNextInspection sets next_inspection_at for each of the given unit IDs.
func (r *Repository) BulkUpdateNextInspection(ctx context.Context, ids []string, nextInspectionAt *int64) error {
	for _, id := range ids {
		if err := r.serializedItems.UpdateNextInspectionAt(ctx, genserialized.UpdateNextInspectionAtParams{
			NextInspectionAt: database.NullInt64Ptr(nextInspectionAt),
			ID:               id,
		}); err != nil {
			return fmt.Errorf("BulkUpdateNextInspection: %w", database.NormalizeError(err))
		}
	}
	return nil
}

// DeleteUnit removes a serialized unit by ID.
func (r *Repository) DeleteUnit(ctx context.Context, id string) error {
	if err := r.serializedItems.Delete(ctx, id); err != nil {
		return fmt.Errorf("DeleteUnit: %w", database.NormalizeError(err))
	}
	return nil
}

// Export returns all non-archived equipment for orgID as a flat list of ExportRows,
// one row per bulk item and one row per serialized unit.
func (r *Repository) Export(ctx context.Context, orgID string) ([]ExportRow, error) {
	rows, err := r.equipment.Export(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("Export: %w", err)
	}
	result := make([]ExportRow, 0, len(rows))
	for _, row := range rows {
		base := ExportRow{
			Name:          row.Name,
			EquipmentType: equipmentTypeLabel(row.EquipmentType),
			TrackingType:  trackingLabel(row.TrackingType),
			Usage:         usageLabel(row.UsageType),
			Category:      row.Category,
			Manufacturer:  row.Manufacturer,
			Location:      row.Location,
			RentalPrice:   row.RentalPrice.String(),
			ResalePrice:   row.ResalePrice.String(),
			Notes:         database.String(row.Notes),
			WeightKg:      row.WeightG.String(),
			WidthCm:       row.WidthMm.String(),
			HeightCm:      row.HeightMm.String(),
			DepthCm:       row.DepthMm.String(),
			VoltageV:      row.VoltageMv.String(),
			CurrentA:      row.CurrentMa.String(),
			PowerW:        row.PowerMw.String(),
			WireGauge:     row.WireGaugeMm2X100.String(),
		}
		switch tracking.Parse(row.TrackingType) {
		case tracking.Bulk:
			bulk, err := r.bulkItems.GetByEquipmentID(ctx, row.ID)
			if err != nil {
				return nil, fmt.Errorf("Export: bulk items for %s: %w", row.ID, err)
			}
			base.Quantity = formatQuantity(bulk.Quantity)
			result = append(result, base)
		case tracking.Serialized:
			serialized, err := r.serializedItems.ListByEquipmentID(ctx, row.ID)
			if err != nil {
				return nil, fmt.Errorf("Export: serialized items for %s: %w", row.ID, err)
			}
			for _, u := range serialized {
				r := base
				r.UnitSerialNumber = u.SerialNumber
				r.UnitManufacturerSerial = database.String(u.ManufacturerSerial)
				r.UnitPurchasePrice = database.NullAs[units.Cents](u.PurchasePrice).String()
				r.UnitPurchasedAt = formatUnixDate(database.Int64Ptr(u.PurchasedAt))
				r.NextInspectionAt = formatUnixDate(database.Int64Ptr(u.NextInspectionAt))
				r.UnitActive = formatActive(u.IsActive)
				r.UnitRemark = database.String(u.Remark)
				result = append(result, r)
			}
		}
	}
	return result, nil
}

// ListContent returns all content items for the equipment definition with equipmentID.
func (r *Repository) ListContent(ctx context.Context, equipmentID string) ([]ContentItem, error) {
	rows, err := r.content.ListByEquipmentID(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("ListContent: %w", err)
	}
	items := make([]ContentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ContentItem{
			ID:          row.ID,
			EquipmentID: row.EquipmentID,
			MemberID:    row.MemberEquipmentID,
			MemberName:  row.MemberName,
			MemberType:  tracking.Type(row.MemberTrackingTypeID.Int64),
			Quantity:    row.Quantity,
		})
	}
	return items, nil
}

// AssignContent inserts a new content entry. Returns database.ErrUniqueConstraint when
// the member is already assigned to this equipment definition.
func (r *Repository) AssignContent(ctx context.Context, a AssignContent) (*ContentItem, error) {
	row, err := r.content.Create(ctx, gencontent.CreateParams{
		ID:                a.ID,
		EquipmentID:       a.EquipmentID,
		MemberEquipmentID: a.MemberID,
		Quantity:          a.Quantity,
	})
	if err != nil {
		normalized := database.NormalizeError(err)
		if errors.Is(normalized, database.ErrUniqueConstraint) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("AssignContent: %w", normalized)
	}
	return &ContentItem{
		ID:          row.ID,
		EquipmentID: row.EquipmentID,
		MemberID:    row.MemberEquipmentID,
		Quantity:    row.Quantity,
	}, nil
}

// RemoveContent deletes the content entry with id.
func (r *Repository) RemoveContent(ctx context.Context, id string) error {
	if err := r.content.Delete(ctx, id); err != nil {
		return fmt.Errorf("RemoveContent: %w", database.NormalizeError(err))
	}
	return nil
}

// ListContainersByMemberID returns all container equipment definitions that include
// memberID in their content definition, ordered by name.
func (r *Repository) ListContainersByMemberID(ctx context.Context, memberID string) ([]PartOf, error) {
	rows, err := r.content.ListContainersByMemberID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("ListContainersByMemberID: %w", err)
	}
	items := make([]PartOf, 0, len(rows))
	for _, row := range rows {
		items = append(items, PartOf{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// Stats returns dashboard aggregate counts and totals for the given org.
func (r *Repository) Stats(ctx context.Context, orgID string) (DashboardStats, error) {
	row, err := r.equipment.Stats(ctx, orgID)
	if err != nil {
		return DashboardStats{}, fmt.Errorf("Stats: %w", err)
	}
	return DashboardStats{
		TotalValue:           units.Cents(row.TotalValue),
		TotalStock:           row.TotalStock,
		EquipmentOverdue:     row.EquipmentOverdue,
		EquipmentOverdueSoon: row.EquipmentOverdueSoon,
	}, nil
}
