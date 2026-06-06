package orgs

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
)

// Options holds configuration for the org service.
type Options struct {
	MaxOrgs int
}

// Service implements business logic for orgs.
type Service struct {
	repo *Repository
	opts Options
}

// NewService returns a new Service backed by repo with the given options.
func NewService(repo *Repository, opts Options) *Service {
	return &Service{repo: repo, opts: opts}
}

// MaxOrgs returns the configured maximum number of allowed orgs.
func (s *Service) MaxOrgs() int {
	return s.opts.MaxOrgs
}

// GetAll returns all orgs.
func (s *Service) GetAll(ctx context.Context) ([]Org, error) {
	orgs, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orgs: %w", err)
	}
	return orgs, nil
}

// GetByID returns the org with the given ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Org, error) {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get org by id: %w", err)
	}
	return org, nil
}

// Update updates the name of the org with the given ID.
func (s *Service) Update(ctx context.Context, u UpdateOrg) (*Org, error) {
	org, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update org: %w", err)
	}
	return org, nil
}

// Delete removes the org with the given ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete org: %w", err)
	}
	return nil
}

// Create creates a new org, enforcing the configured MaxOrgs limit.
func (s *Service) Create(ctx context.Context, createOrg CreateOrg) (*Org, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create org: %w", err)
	}
	if count >= int64(s.opts.MaxOrgs) {
		return nil, fmt.Errorf("failed to create org: %w", database.ErrLimitExceeded)
	}

	org, err := s.repo.Create(ctx, createOrg)
	if err != nil {
		return nil, fmt.Errorf("failed to create org: %w", err)
	}
	return org, nil
}
