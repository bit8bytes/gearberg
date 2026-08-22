// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package settings provides settings functionality.
package settings

import (
	"context"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/uid"
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
func (s *Service) Create(ctx context.Context, orgID string) (*OrgSettings, error) {
	settings, err := s.repo.Create(ctx, createParams{
		ID:    uid.New(),
		OrgID: orgID,
	})
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
