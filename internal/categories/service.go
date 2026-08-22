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

// Package categories provides categories functionality.
package categories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bit8bytes/gearberg/internal/uid"
)

var (
	// ErrNotFound is returned when an equipment category does not exist.
	ErrNotFound = errors.New("equipment category not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("equipment category already exists")
	// ErrLimitExceeded is returned when an org's category limit is reached.
	ErrLimitExceeded = errors.New("equipment category limit exceeded")
	// ErrInUse is returned when a category cannot be deleted due to assigned inventory items.
	ErrInUse = errors.New("equipment category is in use and cannot be deleted")
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

// Category represents an equipment category belonging to a org.
type Category struct {
	ID        string
	OrgID     string
	Name      string
	UpdatedAt int64
	CreatedAt int64
}

// List returns all equipmentCategories belonging to orgID.
func (s *Service) List(ctx context.Context, orgID string) ([]Category, error) {
	equipmentCategories, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment equipmentCategories: %w", err)
	}
	return equipmentCategories, nil
}

// Get returns the category with id.
func (s *Service) Get(ctx context.Context, id string) (*Category, error) {
	category, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment category: %w", err)
	}
	return category, nil
}

// CreateEquipmentCategory holds the data required to create a new equipment category.
type CreateEquipmentCategory struct {
	OrgID string
	Name  string
}

// Create creates a new equipment category, enforcing the configured MaxCategories limit per org.
func (s *Service) Create(ctx context.Context, c CreateEquipmentCategory) (*Category, error) {
	count, err := s.repo.Count(ctx, c.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to create equipment category: %w", err)
	}
	if count >= int64(s.opts.MaxCategories) {
		return nil, ErrLimitExceeded
	}

	category, err := s.repo.Create(ctx, createParams{
		ID:    uid.New(),
		OrgID: c.OrgID,
		Name:  c.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create equipment category: %w", err)
	}
	return category, nil
}

// UpdateEquipmentCategory holds the data required to update an equipment category.
type UpdateEquipmentCategory struct {
	ID   string
	Name string
}

// Update updates the name of an equipment category.
func (s *Service) Update(ctx context.Context, u UpdateEquipmentCategory) (*Category, error) {
	category, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update equipment category: %w", err)
	}
	return category, nil
}

// Upsert returns the ID of the category with the given name within orgID,
// creating it if it does not exist. Bypasses the MaxCategories limit check
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
	c, err := s.repo.GetByName(ctx, orgID, name)
	if err != nil {
		return "", fmt.Errorf("Upsert: %w", err)
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
