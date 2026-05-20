package categories

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
)

// Options holds configuration for the equipment category service.
type Options struct {
	MaxCategories int
}

// Service implements business logic for equipment equipmentCategories.
type Service struct {
	repo *Repository
	opts Options
}

// NewService returns a new Service backed by repo with the given options.
func NewService(repo *Repository, opts Options) *Service {
	return &Service{repo: repo, opts: opts}
}

// MaxCategories returns the configured maximum number of categories per company.
func (s *Service) MaxCategories() int {
	return s.opts.MaxCategories
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

// Create creates a new equipment category, enforcing the configured MaxCategories limit per company.
func (s *Service) Create(ctx context.Context, c CreateEquipmentCategory) (*EquipmentCategory, error) {
	count, err := s.repo.Count(ctx, c.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to create equipment category: %w", err)
	}
	if count >= int64(s.opts.MaxCategories) {
		return nil, fmt.Errorf("failed to create equipment category: %w", database.ErrLimitExceeded)
	}

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
