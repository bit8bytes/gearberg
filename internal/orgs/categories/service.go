package categories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/segmentio/ksuid"
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

// MaxCategories returns the configured maximum number of categories per org.
func (s *Service) MaxCategories() int {
	return s.opts.MaxCategories
}

// GetByOrgID returns all equipmentCategories belonging to orgID.
func (s *Service) GetByOrgID(ctx context.Context, orgID string) ([]EquipmentCategory, error) {
	equipmentCategories, err := s.repo.GetByOrgID(ctx, orgID)
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

// Create creates a new equipment category, enforcing the configured MaxCategories limit per org.
func (s *Service) Create(ctx context.Context, c CreateEquipmentCategory) (*EquipmentCategory, error) {
	count, err := s.repo.Count(ctx, c.OrgID)
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

// EnsureByName returns the ID of the category with the given name within orgID,
// creating it if it does not exist. Bypasses the MaxCategories limit check
// since this is an implicit creation triggered by the user typing a new name.
func (s *Service) EnsureByName(ctx context.Context, orgID, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("EnsureByName: name is blank")
	}
	_, err := s.repo.Create(ctx, CreateEquipmentCategory{
		ID:    ksuid.New().String(),
		OrgID: orgID,
		Name:  name,
	})
	if err != nil && !errors.Is(err, database.ErrUniqueConstraint) {
		return "", fmt.Errorf("EnsureByName: %w", err)
	}
	c, err := s.repo.GetByName(ctx, orgID, name)
	if err != nil {
		return "", fmt.Errorf("EnsureByName: %w", err)
	}
	return c.ID, nil
}

// Delete removes a category. Returns database.ErrForeignKeyViolation when
// inventory items are still assigned to it.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete equipment category: %w", err)
	}
	return nil
}
