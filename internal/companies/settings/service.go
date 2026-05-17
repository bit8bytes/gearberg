package settings

import (
	"context"
	"fmt"
)

// Service implements business logic for company settings.
type Service struct {
	repo *Repository
}

// NewService returns a new Service backed by repo.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetByCompanyID returns the settings for companyID, or nil when none exist yet.
func (s *Service) GetByCompanyID(ctx context.Context, companyID string) (*CompanySettings, error) {
	settings, err := s.repo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company settings: %w", err)
	}
	return settings, nil
}

// Upsert creates or updates the settings for the company identified by u.CompanyID.
func (s *Service) Upsert(ctx context.Context, u UpsertCompanySettings) (*CompanySettings, error) {
	settings, err := s.repo.Upsert(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert company settings: %w", err)
	}
	return settings, nil
}
