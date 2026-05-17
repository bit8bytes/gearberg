package categories

import (
	"context"
	"fmt"
)

// Service implements business logic for equipment equipmentCategories.
type Service struct {
	repo *Repository
}

// NewService returns a new Service backed by repo.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetByCompanyID returns all equipmentCategories belonging to companyID.
func (s *Service) GetByCompanyID(ctx context.Context, companyID string) ([]EquipmentCategory, error) {
	equipmentCategories, err := s.repo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment equipmentCategories: %w", err)
	}
	return equipmentCategories, nil
}

// GetByID returns the category with id.
func (s *Service) GetByID(ctx context.Context, id string) (*EquipmentCategory, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment category: %w", err)
	}
	return category, nil
}

// Create creates a new equipment category.
func (s *Service) Create(ctx context.Context, c CreateEquipmentCategory) (*EquipmentCategory, error) {
	category, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create equipment category: %w", err)
	}
	return category, nil
}

// Update updates the name of an equipment category.
func (s *Service) Update(ctx context.Context, u UpdateEquipmentCategory) (*EquipmentCategory, error) {
	category, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update equipment category: %w", err)
	}
	return category, nil
}

// Delete removes a category. Returns database.ErrForeignKeyViolation when
// inventory items are still assigned to it.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete equipment category: %w", err)
	}
	return nil
}
