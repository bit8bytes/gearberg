package manufacturers

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
)

// Options holds configuration for the manufacturer service.
type Options struct {
	MaxManufacturers int
}

// Service implements business logic for manufacturers.
type Service struct {
	repo *Repository
	opts Options
}

// NewService returns a new Service backed by repo with the given options.
func NewService(repo *Repository, opts Options) *Service {
	return &Service{repo: repo, opts: opts}
}

// MaxManufacturers returns the configured maximum number of manufacturers per org.
func (s *Service) MaxManufacturers() int {
	return s.opts.MaxManufacturers
}

// GetByOrgID returns all manufacturers belonging to orgID.
func (s *Service) GetByOrgID(ctx context.Context, orgID string) ([]Manufacturer, error) {
	manufacturers, err := s.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manufacturers: %w", err)
	}
	return manufacturers, nil
}

// GetByID returns the manufacturer with id.
func (s *Service) GetByID(ctx context.Context, id string) (*Manufacturer, error) {
	manufacturer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get manufacturer: %w", err)
	}
	return manufacturer, nil
}

// Create creates a new manufacturer, enforcing the configured MaxManufacturers limit per org.
func (s *Service) Create(ctx context.Context, c CreateManufacturer) (*Manufacturer, error) {
	count, err := s.repo.Count(ctx, c.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to create manufacturer: %w", err)
	}
	if count >= int64(s.opts.MaxManufacturers) {
		return nil, fmt.Errorf("failed to create manufacturer: %w", database.ErrLimitExceeded)
	}

	manufacturer, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to create manufacturer: %w", err)
	}
	return manufacturer, nil
}

// Update updates the name of a manufacturer.
func (s *Service) Update(ctx context.Context, u UpdateManufacturer) (*Manufacturer, error) {
	manufacturer, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update manufacturer: %w", err)
	}
	return manufacturer, nil
}

// Delete removes a manufacturer. Returns database.ErrForeignKeyViolation when
// inventory items are still assigned to it.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete manufacturer: %w", err)
	}
	return nil
}
