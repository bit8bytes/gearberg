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
package locations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/segmentio/ksuid"
)

// Options holds configuration for the location service.
type Options struct {
	MaxLocations int
}

// Service implements business logic for locations.
type Service struct {
	repo *Repository
	opts Options
}

// NewService returns a new Service backed by repo with the given options.
func NewService(repo *Repository, opts Options) *Service {
	return &Service{repo: repo, opts: opts}
}

// MaxLocations returns the configured maximum number of locations per org.
func (s *Service) MaxLocations() int {
	return s.opts.MaxLocations
}

// GetByOrgID returns all locations belonging to orgID.
func (s *Service) GetByOrgID(ctx context.Context, orgID string) ([]Location, error) {
	locs, err := s.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("GetByOrgID: %w", err)
	}
	return locs, nil
}

// GetByID returns the location with id.
func (s *Service) GetByID(ctx context.Context, id string) (*Location, error) {
	loc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return loc, nil
}

// Create creates a new location, enforcing the configured MaxLocations limit per org.
func (s *Service) Create(ctx context.Context, c CreateLocation) (*Location, error) {
	count, err := s.repo.Count(ctx, c.OrgID)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	if count >= int64(s.opts.MaxLocations) {
		return nil, ErrLimitExceeded
	}

	if c.ID == "" {
		c.ID = ksuid.New().String()
	}

	loc, err := s.repo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return loc, nil
}

// Update updates the name of a location.
func (s *Service) Update(ctx context.Context, u UpdateLocation) (*Location, error) {
	loc, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("Update: %w", err)
	}
	return loc, nil
}

// Delete removes a location. Equipment items referencing it will have their location cleared.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}

// Upsert returns the ID of the location with the given name within orgID,
// creating it if it does not exist. Bypasses the MaxLocations limit check
// since this is an implicit creation triggered by the user typing a new name.
func (s *Service) Upsert(ctx context.Context, orgID, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("Upsert: name is blank")
	}
	_, err := s.repo.Create(ctx, CreateLocation{
		ID:    ksuid.New().String(),
		OrgID: orgID,
		Name:  name,
	})
	if err != nil && !errors.Is(err, ErrConflict) {
		return "", fmt.Errorf("Upsert: %w", err)
	}
	loc, err := s.repo.GetByName(ctx, orgID, name)
	if err != nil {
		return "", fmt.Errorf("Upsert: %w", err)
	}
	return loc.ID, nil
}
