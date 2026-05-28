package inventory

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/bit8bytes/gearberg/internal/companies/categories"
)

// CategoryLister fetches equipment categories by company.
type CategoryLister interface {
	GetByCompanyID(ctx context.Context, companyID string) ([]categories.EquipmentCategory, error)
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

// GetFiltered returns inventory items for companyID, narrowed to the given
// category names and sorted by sortBy ("name", "category", "stock") in sortDir
// ("asc", "desc"). When query is non-empty, only items whose name contains the
// query string (case-insensitive) are returned. Empty values leave results
// unfiltered / in DB order.
func (s *Service) GetFiltered(ctx context.Context, companyID string, categoryNames []string, sortBy, sortDir, query string) ([]Inventory, error) {
	items, err := s.repo.List(ctx, companyID, query)
	if err != nil {
		return nil, fmt.Errorf("GetFiltered: %w", err)
	}
	filtered := items
	if len(categoryNames) > 0 {
		keep := make(map[string]struct{}, len(categoryNames))
		for _, c := range categoryNames {
			keep[c] = struct{}{}
		}
		filtered = items[:0]
		for _, item := range items {
			if _, ok := keep[item.CategoryName]; ok {
				filtered = append(filtered, item)
			}
		}
	}
	sortItems(filtered, sortBy, sortDir)
	return filtered, nil
}

func sortItems(items []Inventory, sortBy, sortDir string) {
	switch sortBy {
	case "name":
		slices.SortFunc(items, func(a, b Inventory) int { return strings.Compare(a.Name, b.Name) })
	case "category":
		slices.SortFunc(items, func(a, b Inventory) int { return strings.Compare(a.CategoryName, b.CategoryName) })
	case "stock":
		slices.SortFunc(items, func(a, b Inventory) int { return cmp.Compare(a.TotalStock, b.TotalStock) })
	default:
		return
	}
	if sortDir == "desc" {
		slices.Reverse(items)
	}
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

// ListCategories returns all equipment categories for companyID.
func (s *Service) ListCategories(ctx context.Context, companyID string) ([]categories.EquipmentCategory, error) {
	cats, err := s.categories.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	return cats, nil
}
