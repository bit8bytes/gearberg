package equipment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genequip "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipment"
	gencontent "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentcombinationitems"
	genidcounter "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentitemidcounter"
	genitems "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentitems"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/segmentio/ksuid"
)

// Repository provides data access for inventory items.
type Repository struct {
	db                     *sql.DB
	equipment              *genequip.Queries
	equipmentItems         *genitems.Queries
	equipmentItemIDCounter *genidcounter.Queries
	content                *gencontent.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:                     db,
		equipment:              genequip.New(db),
		equipmentItems:         genitems.New(db),
		equipmentItemIDCounter: genidcounter.New(db),
		content:                gencontent.New(db),
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
// items whose name contains query (case-insensitive) are returned. When category is
// non-empty, only items in that category are returned. Returns total matching count.
func (r *Repository) List(ctx context.Context, orgID, query, category string, f pagination.Filters) ([]Equipment, int, error) {
	rows, err := r.equipment.List(ctx, genequip.ListParams{
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
	items := make([]Equipment, 0, len(rows))
	for _, row := range rows {
		totalRecords = row.TotalRecords
		items = append(items, Equipment{
			ID:              row.ID,
			OrgID:           row.OrgID,
			Type:            Type(row.TrackingTypeID.Int64),
			UsageType:       UsageType(row.UsageTypeID),
			Name:            row.Name,
			CategoryID:      row.CategoryID,
			CategoryName:    row.CategoryName,
			ManufacturerID:  database.String(row.ManufacturerID),
			LocationID:      database.String(row.LocationID),
			LocationName:    row.LocationName,
			StorageObjectID: database.StringPtr(row.StorageObjectID),
			HasContent:      row.HasContent == 1,
			TotalStock:      row.TotalStock,
			PurchasePrice:   database.Int64Ptr(row.ResalePrice),
			RentalPrice:     database.Int64Ptr(row.RentalPrice),
			Notes:           database.String(row.Notes),
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
		Type:            Type(row.TrackingTypeID.Int64),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		LocationName:    row.LocationName,
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		HasContent:      row.HasContent == 1,
		TotalStock:      row.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.ResalePrice),
		RentalPrice:     database.Int64Ptr(row.RentalPrice),
		Notes:           database.String(row.Notes),
		WeightG:         database.Int64Ptr(row.WeightG),
		WidthMM:         database.Int64Ptr(row.WidthMm),
		HeightMM:        database.Int64Ptr(row.HeightMm),
		DepthMM:         database.Int64Ptr(row.DepthMm),
		PowerMW:         database.Int64Ptr(row.PowerMw),
		CurrentMA:       database.Int64Ptr(row.CurrentMa),
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

// createBulkWith inserts an equipment row and its equipment_items row using the provided queries.
// Callers are responsible for transaction management. cntQ must be scoped to the same transaction
// as eqQ and itemQ to avoid acquiring a second write lock while the first is still held.
func (r *Repository) createBulkWith(ctx context.Context, eqQ *genequip.Queries, itemQ *genitems.Queries, cntQ *genidcounter.Queries, c CreateBulkEquipment) (*Equipment, error) {
	row, err := eqQ.Create(ctx, genequip.CreateParams{
		ID:             c.ID,
		OrgID:          c.OrgID,
		TrackingTypeID: sql.NullInt64{Int64: Bulk.ID(), Valid: true},
		CategoryID:     c.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(c.ManufacturerID)),
		UsageTypeID:    c.UsageTypeID,
		LocationID:     database.NullString(c.LocationID),
		Name:           c.Name,
		HasContent:     c.HasContent,
		RentalPrice:    database.NullInt64Ptr(c.RentalPrice),
		ResalePrice:    database.NullInt64Ptr(c.PurchasePrice),
		Notes:          database.NullString(database.StringOrNil(c.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("createBulkWith: %w", database.NormalizeError(err))
	}
	internalID, err := cntQ.NextInternalID(ctx, c.OrgID)
	if err != nil {
		return nil, fmt.Errorf("createBulkWith: %w", err)
	}
	if _, err := itemQ.Create(ctx, genitems.CreateParams{
		ID:          ksuid.New().String(),
		EquipmentID: row.ID,
		InternalID:  internalID,
		IsActive:    1,
		Quantity:    c.TotalStock,
	}); err != nil {
		return nil, fmt.Errorf("createBulkWith: equipment_items: %w", database.NormalizeError(err))
	}
	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            Type(row.TrackingTypeID.Int64),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      c.TotalStock,
		PurchasePrice:   database.Int64Ptr(row.ResalePrice),
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

// CreateBulk inserts a new bulk inventory item and its equipment_items row atomically.
func (r *Repository) CreateBulk(ctx context.Context, c CreateBulkEquipment) (*Equipment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	m, err := r.createBulkWith(ctx, r.equipment.WithTx(tx), r.equipmentItems.WithTx(tx), r.equipmentItemIDCounter.WithTx(tx), c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("CreateBulk: commit: %w", err)
	}
	return m, nil
}

// CreateBulkTx inserts a new bulk inventory item within an existing transaction.
func (r *Repository) CreateBulkTx(ctx context.Context, tx *sql.Tx, c CreateBulkEquipment) (*Equipment, error) {
	m, err := r.createBulkWith(ctx, r.equipment.WithTx(tx), r.equipmentItems.WithTx(tx), r.equipmentItemIDCounter.WithTx(tx), c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulkTx: %w", err)
	}
	return m, nil
}

// CreateSerialized inserts a serialized inventory item and all its units atomically within tx.
func (r *Repository) CreateSerialized(ctx context.Context, tx *sql.Tx, c CreateSerializedEquipment) (*Equipment, error) {
	eqQ := r.equipment.WithTx(tx)
	itemQ := r.equipmentItems.WithTx(tx)
	row, err := eqQ.Create(ctx, genequip.CreateParams{
		ID:             c.ID,
		OrgID:          c.OrgID,
		TrackingTypeID: sql.NullInt64{Int64: Serialized.ID(), Valid: true},
		CategoryID:     c.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(c.ManufacturerID)),
		UsageTypeID:    c.UsageTypeID,
		LocationID:     database.NullString(c.LocationID),
		Name:           c.Name,
		HasContent:     c.HasContent,
		RentalPrice:    database.NullInt64Ptr(c.RentalPrice),
		ResalePrice:    database.NullInt64Ptr(c.PurchasePrice),
		Notes:          database.NullString(database.StringOrNil(c.Notes)),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateSerialized: %w", database.NormalizeError(err))
	}

	cntQ := r.equipmentItemIDCounter.WithTx(tx)
	for _, u := range c.Units {
		unitInternalID, err := cntQ.NextInternalID(ctx, c.OrgID)
		if err != nil {
			return nil, fmt.Errorf("CreateSerialized: %w", err)
		}
		if _, err := itemQ.Create(ctx, genitems.CreateParams{
			ID:                 u.ID,
			EquipmentID:        row.ID,
			InternalID:         unitInternalID,
			IsActive:           1,
			Quantity:           1,
			ManufacturerSerial: database.NullString(database.StringOrNil(u.SerialNumber)),
		}); err != nil {
			return nil, fmt.Errorf("CreateSerialized: create item: %w", database.NormalizeError(err))
		}
	}

	m := Equipment{
		ID:              row.ID,
		OrgID:           row.OrgID,
		Type:            Type(row.TrackingTypeID.Int64),
		UsageType:       UsageType(row.UsageTypeID),
		Name:            row.Name,
		CategoryID:      row.CategoryID,
		ManufacturerID:  database.String(row.ManufacturerID),
		LocationID:      database.String(row.LocationID),
		StorageObjectID: database.StringPtr(row.StorageObjectID),
		TotalStock:      int64(len(c.Units)),
		PurchasePrice:   database.Int64Ptr(row.ResalePrice),
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
func (r *Repository) UpdateDetails(ctx context.Context, u UpdateEquipmentDetails) error {
	if err := r.equipment.UpdateDetails(ctx, genequip.UpdateDetailsParams{
		ID:             u.ID,
		Name:           u.Name,
		CategoryID:     u.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(u.ManufacturerID)),
		LocationID:     database.NullString(database.StringOrNil(u.LocationID)),
		Notes:          database.NullString(database.StringOrNil(u.Notes)),
	}); err != nil {
		return fmt.Errorf("UpdateDetails: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateDetailsBulk updates the details-tab columns and equipment_items quantity atomically.
func (r *Repository) UpdateDetailsBulk(ctx context.Context, u UpdateEquipmentDetails) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateDetailsBulk: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if err := r.equipment.WithTx(tx).UpdateDetails(ctx, genequip.UpdateDetailsParams{
		ID:             u.ID,
		Name:           u.Name,
		CategoryID:     u.CategoryID,
		ManufacturerID: database.NullString(database.StringOrNil(u.ManufacturerID)),
		LocationID:     database.NullString(database.StringOrNil(u.LocationID)),
		Notes:          database.NullString(database.StringOrNil(u.Notes)),
	}); err != nil {
		return fmt.Errorf("UpdateDetailsBulk: %w", database.NormalizeError(err))
	}
	items, err := r.equipmentItems.WithTx(tx).ListByEquipmentID(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("UpdateDetailsBulk: list items: %w", err)
	}
	var stockItemID string
	for _, item := range items {
		if !item.ParentEquipmentItemID.Valid {
			stockItemID = item.ID
			break
		}
	}
	if stockItemID == "" {
		return fmt.Errorf("UpdateDetailsBulk: no stock item found for equipment %s", u.ID)
	}
	if err := r.equipmentItems.WithTx(tx).SetQuantity(ctx, genitems.SetQuantityParams{
		Quantity: u.TotalStock,
		ID:       stockItemID,
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
		ResalePrice: database.NullInt64Ptr(u.PurchasePrice),
		RentalPrice: database.NullInt64Ptr(u.RentalPrice),
	}); err != nil {
		return fmt.Errorf("UpdatePricing: %w", database.NormalizeError(err))
	}
	return nil
}

// UpdateProperties updates the properties-tab columns for an inventory item.
func (r *Repository) UpdateProperties(ctx context.Context, u UpdateEquipmentProperties) error {
	if err := r.equipment.UpdateProperties(ctx, genequip.UpdatePropertiesParams{
		ID:        u.ID,
		WeightG:   database.NullInt64Ptr(u.WeightG),
		WidthMm:   database.NullInt64Ptr(u.WidthMM),
		HeightMm:  database.NullInt64Ptr(u.HeightMM),
		DepthMm:   database.NullInt64Ptr(u.DepthMM),
		VoltageV:  sql.NullInt64{},
		CurrentMa: database.NullInt64Ptr(u.CurrentMA),
		PowerMw:   database.NullInt64Ptr(u.PowerMW),
	}); err != nil {
		return fmt.Errorf("UpdateProperties: %w", database.NormalizeError(err))
	}
	return nil
}

// ListUnits returns all equipment items for the given inventory item, ordered by serial_number.
func (r *Repository) ListUnits(ctx context.Context, equipmentID string) ([]Unit, error) {
	rows, err := r.equipmentItems.ListByEquipmentID(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}
	us := make([]Unit, 0, len(rows))
	for _, row := range rows {
		us = append(us, Unit{
			ID:                       row.ID,
			EquipmentID:              row.EquipmentID,
			StatusID:                 row.IsActive,
			InternalID:               row.InternalID,
			ManufacturerSerialNumber: database.String(row.ManufacturerSerial),
			Notes:                    database.String(row.Remark),
			Quantity:                 row.Quantity,
			PurchasePrice:            database.Int64Ptr(row.PurchasePrice),
			PurchasedAt:              database.Int64Ptr(row.PurchasedAt),
			LastInspectionAt:         database.Int64Ptr(row.LastInspectedAt),
			CreatedAt:                row.CreatedAt,
			UpdatedAt:                row.UpdatedAt,
		})
	}
	return us, nil
}

// ListAll returns all inventory items for orgID ordered by name, with no pagination.
func (r *Repository) ListAll(ctx context.Context, orgID string) ([]Equipment, error) {
	rows, err := r.equipment.ListAllByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListAll: %w", err)
	}
	items := make([]Equipment, 0, len(rows))
	for _, row := range rows {
		items = append(items, Equipment{
			ID:              row.ID,
			OrgID:           row.OrgID,
			Type:            Type(row.TrackingTypeID.Int64),
			UsageType:       UsageType(row.UsageTypeID),
			Name:            row.Name,
			CategoryID:      row.CategoryID,
			CategoryName:    row.CategoryName,
			ManufacturerID:  database.String(row.ManufacturerID),
			StorageObjectID: database.StringPtr(row.StorageObjectID),
			TotalStock:      row.TotalStock,
			PurchasePrice:   database.Int64Ptr(row.ResalePrice),
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
	if err := r.equipment.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", database.NormalizeError(err))
	}
	return nil
}

// GetUnit returns the unit with id, or database.ErrNotFound when it does not exist.
func (r *Repository) GetUnit(ctx context.Context, id string) (*Unit, error) {
	row, err := r.equipmentItems.GetByID(ctx, id)
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
		InternalID:               row.InternalID,
		ManufacturerSerialNumber: database.String(row.ManufacturerSerial),
		Notes:                    database.String(row.Remark),
		Quantity:                 row.Quantity,
		PurchasePrice:            database.Int64Ptr(row.PurchasePrice),
		PurchasedAt:              database.Int64Ptr(row.PurchasedAt),
		LastInspectionAt:         database.Int64Ptr(row.LastInspectedAt),
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
	}
	return &u, nil
}

// AddUnit inserts a new empty unit for the inventory item.
func (r *Repository) AddUnit(ctx context.Context, orgID string, a AddUnit) (*Unit, error) {
	internalID, err := r.equipmentItemIDCounter.NextInternalID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	row, err := r.equipmentItems.Create(ctx, genitems.CreateParams{
		ID:          a.ID,
		EquipmentID: a.EquipmentID,
		InternalID:  internalID,
		IsActive:    1,
		Quantity:    1,
	})
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", database.NormalizeError(err))
	}
	u := Unit{
		ID:          row.ID,
		EquipmentID: row.EquipmentID,
		StatusID:    row.IsActive,
		InternalID:  row.InternalID,
		CreatedAt:   row.CreatedAt,
	}
	return &u, nil
}

// UpdateUnit updates the editable fields of a unit.
func (r *Repository) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	existing, err := r.equipmentItems.GetByID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.ErrNotFound
		}
		return fmt.Errorf("UpdateUnit: %w", err)
	}
	if err := r.equipmentItems.Update(ctx, genitems.UpdateParams{
		IsActive:           u.StatusID,
		Quantity:           u.Quantity,
		Remark:             database.NullString(database.StringOrNil(u.Notes)),
		PurchasePrice:      database.NullInt64Ptr(u.PurchasePrice),
		PurchasedAt:        database.NullInt64Ptr(u.PurchasedAt),
		LastInspectedAt:    existing.LastInspectedAt,
		ManufacturerSerial: database.NullString(database.StringOrNil(u.ManufacturerSerialNumber)),
		ID:                 u.ID,
	}); err != nil {
		return fmt.Errorf("UpdateUnit: %w", database.NormalizeError(err))
	}
	return nil
}

// DeleteUnit removes a unit by ID.
func (r *Repository) DeleteUnit(ctx context.Context, id string) error {
	if err := r.equipmentItems.Delete(ctx, id); err != nil {
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

// TotalDemandByMemberID returns the total units of memberID committed across all
// combination assignments: sum of (containerTotalStock × perContainerQuantity).
func (r *Repository) TotalDemandByMemberID(ctx context.Context, memberID string) (int64, error) {
	v, err := r.content.TotalDemandByMemberID(ctx, memberID)
	if err != nil {
		return 0, fmt.Errorf("TotalDemandByMemberID: %w", err)
	}
	// sqlc maps COALESCE(SUM(...), 0) to interface{} for SQLite; the driver returns int64.
	if n, ok := v.(int64); ok {
		return n, nil
	}
	return 0, nil
}
