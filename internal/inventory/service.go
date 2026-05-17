package inventory

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/companies/categories"
)

// CategoryLister fetches equipment categories by company.
type CategoryLister interface {
	GetByCompanyID(ctx context.Context, companyID string) ([]categories.EquipmentCategory, error)
}

// Service implements business logic for inventory.
type Service struct {
	categories CategoryLister
}

// NewService returns a new Service.
func NewService(categories CategoryLister) *Service {
	return &Service{categories: categories}
}

// ListCategories returns all equipment categories for companyID.
func (s *Service) ListCategories(ctx context.Context, companyID string) ([]categories.EquipmentCategory, error) {
	cats, err := s.categories.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	return cats, nil
}
