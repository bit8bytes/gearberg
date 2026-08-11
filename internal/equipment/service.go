package equipment

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/segmentio/ksuid"
)

// Manufacturers upserts manufacturers by name.
type Manufacturers interface {
	Upsert(ctx context.Context, orgID, name string) (string, error)
}

// Locations upserts locations by name.
type Locations interface {
	Upsert(ctx context.Context, orgID, name string) (string, error)
}

// Categories upserts categories by name.
type Categories interface {
	Upsert(ctx context.Context, orgID, name string) (string, error)
}

// Service implements business logic for inventory.
type Service struct {
	db            *sql.DB
	repo          *Repository
	manufacturers Manufacturers
	locations     Locations
	categories    Categories
}

// NewService returns a new Service.
func NewService(repo *Repository, db *sql.DB, cats Categories, mfrs Manufacturers, locs Locations) *Service {
	return &Service{db: db, repo: repo, categories: cats, manufacturers: mfrs, locations: locs}
}

// Create creates a new inventory item, dispatching to the correct path based on type.
// Serialized creation runs inside a transaction managed by the service.
// resolveRefs upserts manufacturer and location by name when the corresponding
// ID is empty. It mutates the two ID pointers in place.
func (s *Service) resolveRefs(ctx context.Context, orgID string, manufacturerID *string, manufacturerName string, locationID *string, locationName string, categoryID *string, categoryName string) error {
	if *manufacturerID == "" && manufacturerName != "" {
		id, err := s.manufacturers.Upsert(ctx, orgID, manufacturerName)
		if err != nil {
			return fmt.Errorf("resolveRefs: manufacturer: %w", err)
		}
		*manufacturerID = id
	}
	if *locationID == "" && locationName != "" {
		id, err := s.locations.Upsert(ctx, orgID, locationName)
		if err != nil {
			return fmt.Errorf("resolveRefs: location: %w", err)
		}
		*locationID = id
	}
	if *categoryID == "" && categoryName != "" {
		id, err := s.categories.Upsert(ctx, orgID, categoryName)
		if err != nil {
			return fmt.Errorf("resolveRefs: category: %w", err)
		}
		*categoryID = id
	}
	return nil
}

// Create creates a new equipment item with its associated units.
func (s *Service) Create(ctx context.Context, c CreateEquipment) (*Equipment, error) {
	if err := s.resolveRefs(ctx, c.OrgID, &c.ManufacturerID, c.ManufacturerName, &c.LocationID, c.LocationName, &c.CategoryID, c.CategoryName); err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	if c.TrackingType == Bulk {
		c.HasContent = false
	}
	itemID := ksuid.New().String()
	units := make([]CreateUnit, c.UnitCount)
	for i := range units {
		units[i] = CreateUnit{ID: ksuid.New().String(), OrgID: c.OrgID, EquipmentID: itemID, SerialNumber: serial.New(), IsActive: true}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Both types go through the same tx path; Bulk doesn't need atomicity but a
	// single code path is easier to maintain than two separate non-tx branches.
	var item *Equipment
	switch c.TrackingType {
	case Bulk:
		item, err = s.repo.CreateBulk(ctx, tx, CreateBulkEquipment{ID: itemID, Base: c.Base, TotalStock: c.TotalStock})
	case Serialized:
		item, err = s.repo.CreateSerialized(ctx, tx, CreateSerializedEquipment{ID: itemID, Base: c.Base, Units: units})
	default:
		return nil, fmt.Errorf("Create: unknown equipment type %q", c.TrackingType)
	}
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return item, nil
}

// GetFiltered returns a page of inventory items for orgID, filtered by query,
// category, and archive status, sorted and paginated according to f.
// sortBy may be "code" to order by unit code; otherwise orders by name.
func (s *Service) GetFiltered(ctx context.Context, orgID, query, category string, f pagination.Filters) ([]Equipment, pagination.Metadata, error) {
	items, total, err := s.repo.List(ctx, orgID, query, category, false, f)
	if err != nil {
		return nil, pagination.Metadata{}, fmt.Errorf("GetFiltered: %w", err)
	}
	meta := pagination.CalculateMetadata(total, f.Page, f.PageSize)
	return items, meta, nil
}

// ListAll returns all inventory items for orgID with no pagination, ordered by name.
func (s *Service) ListAll(ctx context.Context, orgID string) ([]Equipment, error) {
	items, err := s.repo.ListAll(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListAll: %w", err)
	}
	return items, nil
}

// ListAllFiltered returns all items for orgID that match the given filters, with no pagination.
func (s *Service) ListAllFiltered(ctx context.Context, orgID, query, category string, showArchived bool) ([]Equipment, error) {
	f := pagination.Filters{Page: 1, PageSize: math.MaxInt32}
	items, _, err := s.repo.List(ctx, orgID, query, category, showArchived, f)
	if err != nil {
		return nil, fmt.Errorf("ListAllFiltered: %w", err)
	}
	return items, nil
}

// findByName returns the first equipment item in items whose name matches name,
// or nil when no match is found.
func findByName(items []Equipment, name string) *Equipment {
	for i := range items {
		if items[i].Name == name {
			return &items[i]
		}
	}
	return nil
}

// AssignContentByName resolves the member by name within the org then assigns it as content.
// Returns ErrNotFound when no equipment matches name.
func (s *Service) AssignContentByName(ctx context.Context, orgID, memberName string, a AssignContent) (*ContentItem, error) {
	all, err := s.repo.ListAll(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("AssignContentByName: %w", err)
	}
	member := findByName(all, memberName)
	if member == nil {
		return nil, fmt.Errorf("AssignContentByName: %w", ErrNotFound)
	}
	a.MemberID = member.ID
	item, err := s.AssignContent(ctx, a, *member)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetByID returns the inventory item with id.
func (s *Service) GetByID(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return item, nil
}

// CreateBulkTx creates a new bulk inventory item within an existing transaction.
func (s *Service) CreateBulkTx(ctx context.Context, tx *sql.Tx, c CreateBulkEquipment) (*Equipment, error) {
	item, err := s.repo.CreateBulk(ctx, tx, c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulkTx: %w", err)
	}
	return item, nil
}

// CreateSerialized creates a new serialized inventory item with all its units in a single transaction.
func (s *Service) CreateSerialized(ctx context.Context, tx *sql.Tx, c CreateSerializedEquipment) (*Equipment, error) {
	item, err := s.repo.CreateSerialized(ctx, tx, c)
	if err != nil {
		return nil, fmt.Errorf("CreateSerialized: %w", err)
	}
	return item, nil
}

// UpdateDetails updates the details-tab fields. For bulk items it also updates
// the stock quantity; for serialized items only the inventory row is written.
func (s *Service) UpdateDetails(ctx context.Context, u UpdateEquipmentDetails) error {
	if err := s.resolveRefs(ctx, u.OrgID, &u.ManufacturerID, u.ManufacturerName, &u.LocationID, u.LocationName, &u.CategoryID, u.CategoryName); err != nil {
		return fmt.Errorf("UpdateDetails: %w", err)
	}
	if u.Type == Bulk {
		if err := s.repo.UpdateDetailsBulk(ctx, u); err != nil {
			return fmt.Errorf("UpdateDetails: %w", err)
		}
		return nil
	}
	if err := s.repo.UpdateDetails(ctx, u); err != nil {
		return fmt.Errorf("UpdateDetails: %w", err)
	}
	return nil
}

// UpdatePricing updates the pricing-tab fields.
func (s *Service) UpdatePricing(ctx context.Context, u UpdateEquipmentPricing) error {
	if err := s.repo.UpdatePricing(ctx, u); err != nil {
		return fmt.Errorf("UpdatePricing: %w", err)
	}
	return nil
}

// UpdateProperties updates the properties-tab fields.
func (s *Service) UpdateProperties(ctx context.Context, u UpdateEquipmentProperties) error {
	if err := s.repo.UpdateProperties(ctx, u); err != nil {
		return fmt.Errorf("UpdateProperties: %w", err)
	}
	return nil
}

// SetImage links or unlinks a storage object from an inventory item.
func (s *Service) SetImage(ctx context.Context, si SetImage) error {
	if err := s.repo.SetImage(ctx, si); err != nil {
		return fmt.Errorf("SetImage: %w", err)
	}
	return nil
}

// Delete removes an inventory item.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}

// Archive sets the archived status of an equipment item.
func (s *Service) Archive(ctx context.Context, a ArchiveEquipment) error {
	if err := s.repo.Archive(ctx, a); err != nil {
		return fmt.Errorf("Archive: %w", err)
	}
	return nil
}

// ListUnits returns all equipment items for the given inventory item, ordered by serial_number.
func (s *Service) ListUnits(ctx context.Context, inventoryID string) ([]Unit, error) {
	units, err := s.repo.ListUnits(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}
	return units, nil
}

// GetUnit returns the unit with id, or database.ErrNotFound when it does not exist.
func (s *Service) GetUnit(ctx context.Context, id string) (*Unit, error) {
	u, err := s.repo.GetUnit(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	return u, nil
}

// AddUnit adds a new unit to the serialized inventory item.
// Returns ErrNotSerializedUnit when the equipment is not serialized.
func (s *Service) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
	item, err := s.repo.GetByID(ctx, a.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	if item.TrackingType != Serialized {
		return nil, fmt.Errorf("AddUnit: %w", ErrNotSerializedUnit)
	}
	a.ID = ksuid.New().String()
	a.SerialNumber = serial.New()
	u, err := s.repo.AddUnit(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	return u, nil
}

// UpdateUnit updates a unit's editable fields.
func (s *Service) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := s.repo.UpdateUnit(ctx, u); err != nil {
		return fmt.Errorf("UpdateUnit: %w", err)
	}
	return nil
}

// BulkUpdateNextInspection sets the next inspection date for multiple units.
func (s *Service) BulkUpdateNextInspection(ctx context.Context, unitIDs []string, nextInspectionAt *int64) error {
	if err := s.repo.BulkUpdateNextInspection(ctx, unitIDs, nextInspectionAt); err != nil {
		return fmt.Errorf("BulkUpdateNextInspection: %w", err)
	}
	return nil
}

// DeleteUnit removes a unit from the serialized inventory item.
func (s *Service) DeleteUnit(ctx context.Context, unitID string) error {
	if err := s.repo.DeleteUnit(ctx, unitID); err != nil {
		return fmt.Errorf("DeleteUnit: %w", err)
	}
	return nil
}

// GetUnitsContainer fetches the equipment item and verifies it supports the units tab.
// Returns ErrNotFound when the item does not exist, ErrNoUnitsTab when it is bulk equipment.
func (s *Service) GetUnitsContainer(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUnitsContainer: %w", err)
	}
	if item.TrackingType != Serialized {
		return nil, fmt.Errorf("GetUnitsContainer: %w", ErrNoUnitsTab)
	}
	return item, nil
}

// GetContentContainer fetches the equipment item and verifies it supports the content tab.
// Returns ErrNotFound when the item does not exist, ErrNoContentTab when it is not a container.
func (s *Service) GetContentContainer(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetContentContainer: %w", err)
	}
	if item.TrackingType != Serialized && !item.HasContent {
		return nil, fmt.Errorf("GetContentContainer: %w", ErrNoContentTab)
	}
	return item, nil
}

// ListContent returns all content items for equipmentID.
func (s *Service) ListContent(ctx context.Context, equipmentID string) ([]ContentItem, error) {
	items, err := s.repo.ListContent(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("ListContent: %w", err)
	}
	return items, nil
}

// AssignContent adds an equipment item as content of a container.
// Returns ErrInvalidContent when the member is the container itself or is itself a container.
// Returns ErrConflict when the member is already assigned.
// Stock sufficiency is not enforced here; callers should surface warnings via ListContent.
func (s *Service) AssignContent(ctx context.Context, a AssignContent, member Equipment) (*ContentItem, error) {
	a.ID = ksuid.New().String()
	if member.ID == a.EquipmentID {
		return nil, fmt.Errorf("AssignContent: %w: an item cannot contain itself", ErrInvalidContent)
	}
	if member.HasContent || member.TrackingType == Serialized {
		return nil, fmt.Errorf("AssignContent: %w: cannot add a container as content", ErrInvalidContent)
	}
	item, err := s.repo.AssignContent(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("AssignContent: %w", err)
	}
	return item, nil
}

// RemoveContent deletes the content entry with id.
func (s *Service) RemoveContent(ctx context.Context, id string) error {
	if err := s.repo.RemoveContent(ctx, id); err != nil {
		return fmt.Errorf("RemoveContent: %w", err)
	}
	return nil
}

// ListContainers returns all container equipment definitions that include memberID
// in their content definition.
func (s *Service) ListContainers(ctx context.Context, memberID string) ([]PartOf, error) {
	items, err := s.repo.ListContainersByMemberID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("ListContainers: %w", err)
	}
	return items, nil
}
