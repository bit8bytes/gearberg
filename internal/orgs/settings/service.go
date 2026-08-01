package settings

import (
	"context"
	"fmt"
)

// Service implements business logic for org settings.
type Service struct {
	repo *Repository
}

// NewService returns a new Service backed by repo.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Get returns the settings for orgID, or nil when none exist yet.
func (s *Service) Get(ctx context.Context, orgID string) (*OrgSettings, error) {
	settings, err := s.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get org settings: %w", err)
	}
	return settings, nil
}

// Create inserts new settings for orgID using the migration defaults.
func (s *Service) Create(ctx context.Context, id, orgID string) (*OrgSettings, error) {
	settings, err := s.repo.Create(ctx, id, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to create org settings: %w", err)
	}
	return settings, nil
}

// Update applies currency, vat_rate, and timezone changes to the settings for u.OrgID.
func (s *Service) Update(ctx context.Context, u UpdateOrgSettings) (*OrgSettings, error) {
	settings, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update org settings: %w", err)
	}
	return settings, nil
}
