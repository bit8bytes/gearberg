package inventory

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/companies/categories"
)

// Lister fetches inventory items by company.
type Lister interface {
	GetByCompanyID(ctx context.Context, companyID string) ([]Inventory, error)
}

// CategoryLister fetches equipment categories by company.
type CategoryLister interface {
	GetByCompanyID(ctx context.Context, companyID string) ([]categories.EquipmentCategory, error)
}

// Service implements business logic for inventory.
type Service struct {
	repo       Lister
	categories CategoryLister
}

// NewService returns a new Service.
func NewService(repo Lister, cats CategoryLister) *Service {
	return &Service{repo: repo, categories: cats}
}

// GetAll returns all inventory items for companyID.
func (s *Service) GetAll(ctx context.Context, companyID string) ([]Inventory, error) {
	items, err := s.repo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}
	return items, nil
}

// ListCategories returns all equipment categories for companyID.
func (s *Service) ListCategories(ctx context.Context, companyID string) ([]categories.EquipmentCategory, error) {
	cats, err := s.categories.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	return cats, nil
}
