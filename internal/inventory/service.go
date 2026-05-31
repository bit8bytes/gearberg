package inventory

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/internal/pagination"
)

// CategoryLister fetches equipment categories by org.
type CategoryLister interface {
	GetByOrgID(ctx context.Context, orgID string) ([]categories.EquipmentCategory, error)
}

// Service implements business logic for inventory.
type Service struct {
	repo       *Repository
	categories CategoryLister
}

// NewService returns a new Service.
func NewService(repo *Repository, cats CategoryLister) *Service {
	return &Service{repo: repo, categories: cats}
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

// GetByID returns the inventory item with id.
func (s *Service) GetByID(ctx context.Context, id string) (*Inventory, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return item, nil
}

// Create creates a new inventory item.
func (s *Service) Create(ctx context.Context, c CreateInventory) (*Inventory, error) {
	item, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
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

// ListCategories returns all equipment categories for orgID.
func (s *Service) ListCategories(ctx context.Context, orgID string) ([]categories.EquipmentCategory, error) {
	cats, err := s.categories.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	return cats, nil
}
