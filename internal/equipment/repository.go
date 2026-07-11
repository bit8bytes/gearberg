package equipment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/bit8bytes/gearberg/internal/database"
	genequip "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipment"
	genbulk "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentbulkitems"
	gencontent "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentcombinationitems"
	genserialized "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentserializeditems"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/segmentio/ksuid"
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
func (r *Repository) List(ctx context.Context, orgID, query, category string, showArchived bool, f pagination.Filters) ([]Equipment, int, error) {
	rows, err := r.equipment.List(ctx, genequip.ListParams{
		OrgID:      orgID,
		NameQuery:  query,
		Category:   category,
		IsArchived: database.Bool(showArchived),
		PageOffset: int64(f.Offset()),
		PageLimit:  int64(f.Limit()),
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
			Kind:            KindFromString(row.EquipmentTypeName),
			Type:            Type(row.TrackingTypeID.Int64),
			UsageType:       UsageType(row.UsageTypeID),
			Name:            row.Name,
			CategoryID:      database.String(row.CategoryID),
			CategoryName:    row.CategoryName,
			ManufacturerID:  database.String(row.ManufacturerID),
			LocationID:      database.String(row.LocationID),
			LocationName:    row.LocationName,
			StorageObjectID: database.StringPtr(row.StorageObjectID),
			TotalStock:      row.TotalStock,
			HasContent:      row.HasContent == 1,
			IsArchived:      row.IsArchived == 1,
			Notes:           database.String(row.Notes),
			Pricing:         pricingFromListRow(row),
			Properties:      propertiesFromListRow(row),
			UpdatedAt:       row.UpdatedAt,
			CreatedAt:       row.CreatedAt,
		})
	}
	return items, int(totalRecords), nil
}

// GetByID returns the inventory item with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetByID(ctx context.Context, id string) (*Equipment, error) {
	row, err := r.equipment.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Kind:            KindFromString(row.EquipmentTypeName),
		Type:            Type(row.TrackingTypeID.Int64),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      database.String(row.CategoryID),
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		LocationName:    row.LocationName,
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      row.TotalStock,
		ContentCount:    row.ContentCount,
		HasContent:      row.HasContent == 1,
		IsArchived:      row.IsArchived == 1,
		Notes:           database.String(row.Notes),
		Pricing:         pricingFromGetByIDRow(row),
		Properties:      propertiesFromGetByIDRow(row),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
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
		EquipmentTypeID:  Physical.ID(),
		TrackingTypeID:   sql.NullInt64{Int64: Bulk.ID(), Valid: true},
		CategoryID:       database.NullString(database.StringOrNil(c.CategoryID)),
		ManufacturerID:   database.NullString(database.StringOrNil(c.ManufacturerID)),
		UsageTypeID:      c.UsageTypeID,
		LocationID:       database.NullString(c.LocationID),
		Name:             c.Name,
		HasContent:       database.Bool(c.HasContent),
		IsArchived:       0,
		RentalPrice:      database.NullOf(c.Pricing.RentalPrice),
		ResalePrice:      database.NullOf(c.Pricing.PurchasePrice),
		Notes:            database.NullString(database.StringOrNil(c.Notes)),
		WeightG:          database.NullOf(c.Properties.Weight),
		WidthMm:          database.NullOf(c.Properties.Width),
		HeightMm:         database.NullOf(c.Properties.Height),
		DepthMm:          database.NullOf(c.Properties.Depth),
		CurrentMa:        database.NullOf(c.Properties.Current),
		PowerMw:          database.NullOf(c.Properties.Power),
		VoltageMv:        database.NullOf(c.Properties.Voltage),
		WireGaugeMm2X100: database.NullOf(c.Properties.WireGauge),
	})
	if err != nil {
		return nil, fmt.Errorf("createBulkWith: %w", database.NormalizeError(err))
	}
	if _, err := bulkQ.Create(ctx, genbulk.CreateParams{
		ID:          ksuid.New().String(),
		EquipmentID: row.ID,
		Quantity:    c.TotalStock,
	}); err != nil {
		return nil, fmt.Errorf("createBulkWith: equipment_bulk_items: %w", database.NormalizeError(err))
	}
	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Kind:            Physical,
		Type:            Type(row.TrackingTypeID.Int64),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      database.String(row.CategoryID),
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      c.TotalStock,
		HasContent:      row.HasContent == 1,
		Notes:           database.String(row.Notes),
		Pricing:         pricingFromCreateRow(row),
		Properties:      propertiesFromCreateRow(row),
		CreatedAt:       row.CreatedAt,
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
		EquipmentTypeID:  Physical.ID(),
		TrackingTypeID:   sql.NullInt64{Int64: Serialized.ID(), Valid: true},
		CategoryID:       database.NullString(database.StringOrNil(c.CategoryID)),
		ManufacturerID:   database.NullString(database.StringOrNil(c.ManufacturerID)),
		UsageTypeID:      c.UsageTypeID,
		LocationID:       database.NullString(c.LocationID),
		Name:             c.Name,
		HasContent:       database.Bool(c.HasContent),
		IsArchived:       0,
		RentalPrice:      database.NullOf(c.Pricing.RentalPrice),
		ResalePrice:      database.NullOf(c.Pricing.PurchasePrice),
		Notes:            database.NullString(database.StringOrNil(c.Notes)),
		WeightG:          database.NullOf(c.Properties.Weight),
		WidthMm:          database.NullOf(c.Properties.Width),
		HeightMm:         database.NullOf(c.Properties.Height),
		DepthMm:          database.NullOf(c.Properties.Depth),
		CurrentMa:        database.NullOf(c.Properties.Current),
		PowerMw:          database.NullOf(c.Properties.Power),
		VoltageMv:        database.NullOf(c.Properties.Voltage),
		WireGaugeMm2X100: database.NullOf(c.Properties.WireGauge),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateSerialized: %w", database.NormalizeError(err))
	}

	for _, u := range c.Units {
		if _, err := itemQ.Create(ctx, genserialized.CreateParams{
			ID:           u.ID,
			OrgID:        c.OrgID,
			EquipmentID:  row.ID,
			SerialNumber: u.SerialNumber,
			Code:         database.NullString(database.StringOrNil(u.Code)),
			IsActive:     1,
		}); err != nil {
			return nil, fmt.Errorf("CreateSerialized: create item: %w", database.NormalizeError(err))
		}
	}

	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Kind:            Physical,
		Type:            Type(row.TrackingTypeID.Int64),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      database.String(row.CategoryID),
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      int64(len(c.Units)),
		HasContent:      row.HasContent == 1,
		Notes:           database.String(row.Notes),
		Pricing:         pricingFromCreateRow(row),
		Properties:      propertiesFromCreateRow(row),
		CreatedAt:       row.CreatedAt,
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
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: commit: %w", err)
	}
	return nil
}

// UpdatePricing updates the pricing-tab columns for an inventory item.
func (r *Repository) UpdatePricing(ctx context.Context, u UpdateEquipmentPricing) error {
	if err := r.equipment.UpdatePricing(ctx, genequip.UpdatePricingParams{
		ID:          u.ID,
		ResalePrice: database.NullOf(u.Pricing.PurchasePrice),
		RentalPrice: database.NullOf(u.Pricing.RentalPrice),
	}); err != nil {
		return fmt.Errorf("UpdatePricing: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateProperties updates the properties-tab columns for an inventory item.
func (r *Repository) UpdateProperties(ctx context.Context, u UpdateEquipmentProperties) error {
	if err := r.equipment.UpdateProperties(ctx, genequip.UpdatePropertiesParams{
		ID:               u.ID,
		WeightG:          database.NullOf(u.Properties.Weight),
		WidthMm:          database.NullOf(u.Properties.Width),
		HeightMm:         database.NullOf(u.Properties.Height),
		DepthMm:          database.NullOf(u.Properties.Depth),
		VoltageMv:        database.NullOf(u.Properties.Voltage),
		CurrentMa:        database.NullOf(u.Properties.Current),
		PowerMw:          database.NullOf(u.Properties.Power),
		WireGaugeMm2X100: database.NullOf(u.Properties.WireGauge),
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
			Code:                     database.String(row.Code),
			ManufacturerSerialNumber: database.String(row.ManufacturerSerial),
			Notes:                    database.String(row.Remark),
			PurchasePrice:            database.NullAs[Cents](row.PurchasePrice),
			PurchasedAt:              database.Int64Ptr(row.PurchasedAt),
			NextInspectionAt:         database.Int64Ptr(row.NextInspectionAt),
			CreatedAt:                row.CreatedAt,
			UpdatedAt:                row.UpdatedAt,
		})
	}
	return us, nil
}

// ListAll returns all inventory items for orgID ordered by name, with no pagination.
// Passes -1 as the is_archived sentinel to skip the archived filter and return all items.
func (r *Repository) ListAll(ctx context.Context, orgID string) ([]Equipment, error) {
	rows, err := r.equipment.List(ctx, genequip.ListParams{
		OrgID:      orgID,
		NameQuery:  "",
		Category:   "",
		IsArchived: -1,
		PageOffset: 0,
		PageLimit:  math.MaxInt32,
	})
	if err != nil {
		return nil, fmt.Errorf("ListAll: %w", err)
	}
	items, _, err := listRowsToEquipment(rows)
	return items, err
}

// Archive sets the is_archived flag on an inventory item.
func (r *Repository) Archive(ctx context.Context, a ArchiveEquipment) error {
	if err := r.equipment.UpdateArchived(ctx, genequip.UpdateArchivedParams{
		ID:         a.ID,
		IsArchived: database.Bool(a.IsArchived),
	}); err != nil {
		return fmt.Errorf("Archive: %w", database.NormalizeError(err))
	}
	return nil
}

// Delete removes the inventory item. Returns database.ErrForeignKeyViolation when
// active rental line items reference it.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.equipment.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

// GetUnit returns the serialized unit with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetUnit(ctx context.Context, id string) (*Unit, error) {
	row, err := r.serializedItems.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	u := Unit{
		ID:                       row.ID,
		EquipmentID:              row.EquipmentID,
		StatusID:                 row.IsActive,
		SerialNumber:             row.SerialNumber,
		Code:                     database.String(row.Code),
		ManufacturerSerialNumber: database.String(row.ManufacturerSerial),
		Notes:                    database.String(row.Remark),
		PurchasePrice:            database.NullAs[Cents](row.PurchasePrice),
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
		Code:         database.NullString(database.StringOrNil(a.Code)),
		IsActive:     1,
	})
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", database.NormalizeError(err))
	}
	u := Unit{
		ID:           row.ID,
		EquipmentID:  row.EquipmentID,
		StatusID:     row.IsActive,
		SerialNumber: row.SerialNumber,
		Code:         database.String(row.Code),
		CreatedAt:    row.CreatedAt,
	}
	return &u, nil
}

// UpdateUnit updates the editable fields of a serialized unit.
func (r *Repository) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := r.serializedItems.Update(ctx, genserialized.UpdateParams{
		Code:               database.NullString(database.StringOrNil(u.Code)),
		IsActive:           u.StatusID,
		Remark:             database.NullString(database.StringOrNil(u.Notes)),
		PurchasePrice:      database.NullOf(u.PurchasePrice),
		PurchasedAt:        database.NullInt64Ptr(u.PurchasedAt),
		NextInspectionAt:   database.NullInt64Ptr(u.NextInspectionAt),
		ManufacturerSerial: database.NullString(database.StringOrNil(u.ManufacturerSerialNumber)),
		ID:                 u.ID,
	}); err != nil {
		return fmt.Errorf("UpdateUnit: %w", database.NormalizeError(err))
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
			MemberType:  Type(row.MemberTrackingTypeID.Int64),
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
		return nil, fmt.Errorf("AssignContent: %w", database.NormalizeError(err))
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
