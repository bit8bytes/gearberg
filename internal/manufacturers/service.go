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

// Package manufacturers provides manufacturers functionality.
package manufacturers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bit8bytes/gearberg/internal/uid"
)

var (
	// ErrNotFound is returned when a manufacturer does not exist.
	ErrNotFound = errors.New("manufacturer not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("manufacturer already exists")
	// ErrLimitExceeded is returned when an org's manufacturer limit is reached.
	ErrLimitExceeded = errors.New("manufacturer limit exceeded")
	// ErrInUse is returned when a manufacturer cannot be deleted due to assigned inventory items.
	ErrInUse = errors.New("manufacturer is in use and cannot be deleted")
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

// Manufacturer represents a manufacturer belonging to an org.
type Manufacturer struct {
	ID        string
	OrgID     string
	Name      string
	UpdatedAt int64
	CreatedAt int64
}

// List returns all manufacturers belonging to orgID.
func (s *Service) List(ctx context.Context, orgID string) ([]Manufacturer, error) {
	manufacturers, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manufacturers: %w", err)
	}
	return manufacturers, nil
}

// Get returns the manufacturer with id.
func (s *Service) Get(ctx context.Context, id string) (*Manufacturer, error) {
	manufacturer, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get manufacturer: %w", err)
	}
	return manufacturer, nil
}

// CreateManufacturer holds the data required to create a new manufacturer.
type CreateManufacturer struct {
	OrgID string
	Name  string
}

// Create creates a new manufacturer, enforcing the configured MaxManufacturers limit per org.
func (s *Service) Create(ctx context.Context, c CreateManufacturer) (*Manufacturer, error) {
	count, err := s.repo.Count(ctx, c.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to create manufacturer: %w", err)
	}
	if count >= int64(s.opts.MaxManufacturers) {
		return nil, ErrLimitExceeded
	}

	manufacturer, err := s.repo.Create(ctx, createParams{
		ID:    uid.New(),
		OrgID: c.OrgID,
		Name:  c.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manufacturer: %w", err)
	}
	return manufacturer, nil
}

// UpdateManufacturer holds the data required to update a manufacturer.
type UpdateManufacturer struct {
	ID   string
	Name string
}

// Update updates the name of a manufacturer.
func (s *Service) Update(ctx context.Context, u UpdateManufacturer) (*Manufacturer, error) {
	manufacturer, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update manufacturer: %w", err)
	}
	return manufacturer, nil
}

// Upsert returns the ID of the manufacturer with the given name within orgID,
// creating it if it does not exist. Bypasses the MaxManufacturers limit check
// since this is an implicit creation triggered by the user typing a new name.
func (s *Service) Upsert(ctx context.Context, orgID, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("Upsert: name is blank")
	}
	_, err := s.repo.Create(ctx, createParams{
		ID:    uid.New(),
		OrgID: orgID,
		Name:  name,
	})
	if err != nil && !errors.Is(err, ErrConflict) {
		return "", fmt.Errorf("Upsert: %w", err)
	}
	m, err := s.repo.GetByName(ctx, orgID, name)
	if err != nil {
		return "", fmt.Errorf("Upsert: %w", err)
	}
	return m.ID, nil
}

// Delete removes a manufacturer. Returns database.ErrForeignKeyViolation when
// inventory items are still assigned to it.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete manufacturer: %w", err)
	}
	return nil
}
