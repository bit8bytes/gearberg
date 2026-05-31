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

// GetByOrgID returns the settings for orgID, or nil when none exist yet.
func (s *Service) GetByOrgID(ctx context.Context, orgID string) (*OrgSettings, error) {
	settings, err := s.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get org settings: %w", err)
	}
	return settings, nil
}

// Upsert creates or updates the settings for the org identified by u.OrgID.
func (s *Service) Upsert(ctx context.Context, u UpsertOrgSettings) (*OrgSettings, error) {
	settings, err := s.repo.Upsert(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert org settings: %w", err)
	}
	return settings, nil
}
