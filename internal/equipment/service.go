package equipment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/equipment/categories"
	"github.com/bit8bytes/gearberg/internal/equipment/locations"
	"github.com/bit8bytes/gearberg/internal/equipment/manufacturers"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/segmentio/ksuid"
)

// CategoryLister fetches equipment categories by org.
type CategoryLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]categories.EquipmentCategory, error)
}

// ManufacturerLister fetches manufacturers by org.
type ManufacturerLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]manufacturers.Manufacturer, error)
}

// LocationLister fetches locations by org.
type LocationLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]locations.Location, error)
}

// Service implements business logic for inventory.
type Service struct {
	db            *sql.DB
	repo          *Repository
	categories    CategoryLister
	manufacturers ManufacturerLister
	locations     LocationLister
}

// NewService returns a new Service.
func NewService(repo *Repository, db *sql.DB, cats CategoryLister, mfrs ManufacturerLister, locs LocationLister) *Service {
	return &Service{db: db, repo: repo, categories: cats, manufacturers: mfrs, locations: locs}
}

// Create creates a new inventory item, dispatching to the correct path based on type.
// Serialized creation runs inside a transaction managed by the service.
func (s *Service) Create(ctx context.Context, c CreateEquipment) (*Equipment, error) {
	if c.Type == Bulk {
		return s.CreateBulk(ctx, CreateBulkEquipment{Base: c.Base, TotalStock: c.TotalStock})
	}
	units := make([]CreateUnit, c.UnitCount)
	for i := range units {
		units[i] = CreateUnit{ID: ksuid.New().String(), OrgID: c.OrgID, EquipmentID: c.ID, SerialNumber: serial.New()}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	item, err := s.repo.CreateSerialized(ctx, tx, CreateSerializedEquipment{Base: c.Base, Units: units})
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
func (s *Service) GetFiltered(ctx context.Context, orgID, query, category string, showArchived bool, f pagination.Filters) ([]Equipment, pagination.Metadata, error) {
	items, total, err := s.repo.List(ctx, orgID, query, category, showArchived, f)
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

// GetByID returns the inventory item with id.
func (s *Service) GetByID(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return item, nil
}

// CreateBulk creates a new bulk inventory item with an auto-assigned code.
func (s *Service) CreateBulk(ctx context.Context, c CreateBulkEquipment) (*Equipment, error) {
	item, err := s.repo.CreateBulk(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", err)
	}
	return item, nil
}

// CreateBulkTx creates a new bulk inventory item within an existing transaction.
func (s *Service) CreateBulkTx(ctx context.Context, tx *sql.Tx, c CreateBulkEquipment) (*Equipment, error) {
	item, err := s.repo.CreateBulkTx(ctx, tx, c)
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

// ListCategories returns all equipment categories for orgID.
func (s *Service) ListCategories(ctx context.Context, orgID string) ([]categories.EquipmentCategory, error) {
	cats, err := s.categories.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	return cats, nil
}

// ListManufacturers returns all manufacturers for orgID.
func (s *Service) ListManufacturers(ctx context.Context, orgID string) ([]manufacturers.Manufacturer, error) {
	mfrs, err := s.manufacturers.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListManufacturers: %w", err)
	}
	return mfrs, nil
}

// ListLocations returns all locations for orgID.
func (s *Service) ListLocations(ctx context.Context, orgID string) ([]locations.Location, error) {
	locs, err := s.locations.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListLocations: %w", err)
	}
	return locs, nil
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
func (s *Service) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
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

// DeleteUnit removes a unit from the serialized inventory item.
func (s *Service) DeleteUnit(ctx context.Context, unitID string) error {
	if err := s.repo.DeleteUnit(ctx, unitID); err != nil {
		return fmt.Errorf("DeleteUnit: %w", err)
	}
	return nil
}

// ListContent returns all content items for equipmentID.
func (s *Service) ListContent(ctx context.Context, equipmentID string) ([]ContentItem, error) {
	items, err := s.repo.ListContent(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("ListContent: %w", err)
	}
	return items, nil
}

// AssignContent adds an equipment definition as content of a container. Returns
// database.ErrUniqueConstraint when already assigned, or an error when memberID == equipmentID.
// Stock sufficiency is not enforced here; callers should surface warnings via ListContent.
func (s *Service) AssignContent(ctx context.Context, a AssignContent) (*ContentItem, error) {
	if a.MemberID == a.EquipmentID {
		return nil, fmt.Errorf("AssignContent: an item cannot contain itself")
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
