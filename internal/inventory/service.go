package inventory

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/internal/orgs/manufacturers"
	"github.com/bit8bytes/gearberg/internal/pagination"
)

// CategoryLister fetches equipment categories by org.
type CategoryLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]categories.EquipmentCategory, error)
}

// ManufacturerLister fetches manufacturers by org.
type ManufacturerLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]manufacturers.Manufacturer, error)
}

// Service implements business logic for inventory.
type Service struct {
	repo          *Repository
	categories    CategoryLister
	manufacturers ManufacturerLister
}

// NewService returns a new Service.
func NewService(repo *Repository, cats CategoryLister, mfrs ManufacturerLister) *Service {
	return &Service{repo: repo, categories: cats, manufacturers: mfrs}
}

// GetFiltered returns a page of inventory items for orgID, filtered by query
// and category, sorted and paginated according to f.
func (s *Service) GetFiltered(ctx context.Context, orgID, query, category string, f pagination.Filters) ([]Inventory, pagination.Metadata, error) {
	items, total, err := s.repo.List(ctx, orgID, query, category, f)
	if err != nil {
		return nil, pagination.Metadata{}, fmt.Errorf("GetFiltered: %w", err)
	}
	meta := pagination.CalculateMetadata(total, f.Page, f.PageSize)
	return items, meta, nil
}

// ListAll returns all inventory items for orgID with no pagination, ordered by name.
func (s *Service) ListAll(ctx context.Context, orgID string) ([]Inventory, error) {
	items, err := s.repo.ListAll(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListAll: %w", err)
	}
	return items, nil
}

// GetByID returns the inventory item with id.
func (s *Service) GetByID(ctx context.Context, id string) (*Inventory, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return item, nil
}

// CreateBulk creates a new bulk inventory item with an auto-assigned code.
func (s *Service) CreateBulk(ctx context.Context, c CreateBulkInventory) (*Inventory, error) {
	item, err := s.repo.CreateBulk(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulk: %w", err)
	}
	return item, nil
}

// CreateBulkTx creates a new bulk inventory item within an existing transaction.
func (s *Service) CreateBulkTx(ctx context.Context, tx *sql.Tx, c CreateBulkInventory) (*Inventory, error) {
	item, err := s.repo.CreateBulkTx(ctx, tx, c)
	if err != nil {
		return nil, fmt.Errorf("CreateBulkTx: %w", err)
	}
	return item, nil
}

// UpdateTx updates an inventory item within an existing transaction.
func (s *Service) UpdateTx(ctx context.Context, tx *sql.Tx, u UpdateInventory) (*Inventory, error) {
	item, err := s.repo.UpdateTx(ctx, tx, u)
	if err != nil {
		return nil, fmt.Errorf("UpdateTx: %w", err)
	}
	return item, nil
}

// CreateSerialized creates a new serialized inventory item with all its units in a single transaction.
func (s *Service) CreateSerialized(ctx context.Context, tx *sql.Tx, c CreateSerializedInventory) (*Inventory, error) {
	item, err := s.repo.CreateSerialized(ctx, tx, c)
	if err != nil {
		return nil, fmt.Errorf("CreateSerialized: %w", err)
	}
	return item, nil
}

// Update updates an inventory item.
func (s *Service) Update(ctx context.Context, u UpdateInventory) (*Inventory, error) {
	item, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("Update: %w", err)
	}
	return item, nil
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

// ListUnits returns all inventory units for the given inventory item, ordered by unit_number.
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

// GetUnit returns the unit with id, or database.ErrNotFound when it does not exist.
func (s *Service) GetUnit(ctx context.Context, id string) (*Unit, error) {
	u, err := s.repo.GetUnit(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	return u, nil
}

// AddUnit adds a new empty unit to the serialized inventory item.
// It also increments total_stock on the inventory record.
func (s *Service) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
	u, err := s.repo.AddUnit(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	if err := s.repo.UpdateTotalStock(ctx, a.InventoryID, 1); err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	return u, nil
}

// ListUnitStatuses returns all rows from the unit_statuses lookup table.
func (s *Service) ListUnitStatuses(ctx context.Context) ([]UnitStatusEntry, error) {
	statuses, err := s.repo.ListUnitStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListUnitStatuses: %w", err)
	}
	return statuses, nil
}

// UpdateUnit updates a unit's editable fields.
func (s *Service) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := s.repo.UpdateUnit(ctx, u); err != nil {
		return fmt.Errorf("UpdateUnit: %w", err)
	}
	return nil
}

// DeleteUnit removes a unit and decrements total_stock on the inventory record.
func (s *Service) DeleteUnit(ctx context.Context, unitID, inventoryID string) error {
	if err := s.repo.DeleteUnit(ctx, unitID); err != nil {
		return fmt.Errorf("DeleteUnit: %w", err)
	}
	if err := s.repo.UpdateTotalStock(ctx, inventoryID, -1); err != nil {
		return fmt.Errorf("DeleteUnit: %w", err)
	}
	return nil
}
