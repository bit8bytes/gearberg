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

// Package equipment provides equipment functionality.
package equipment

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/bit8bytes/gearberg/internal/uid"
)

// Service implements business logic for inventory.
type Service struct {
	db   *sql.DB
	repo *Repository
}

// NewService returns a new Service.
func NewService(repo *Repository, db *sql.DB) *Service {
	return &Service{db: db, repo: repo}
}

// Create creates a new inventory item, dispatching to the correct path based on type.
// Serialized creation runs inside a transaction managed by the service.
func (s *Service) Create(ctx context.Context, c CreateEquipment) (*Equipment, error) {
	var trackingType tracking.Type
	switch c.ItemType {
	case tracking.Bulk.String():
		trackingType = tracking.Bulk
		c.EquipmentType = StandardType
	case tracking.Serialized.String():
		trackingType = tracking.Serialized
		c.EquipmentType = StandardType
	case KitType.String():
		trackingType = tracking.Serialized
		c.EquipmentType = KitType
	default:
		return nil, fmt.Errorf("Create: unknown item type %q", c.ItemType)
	}

	itemID := uid.New()
	units := make([]CreateUnit, c.UnitCount)
	for i := range units {
		units[i] = CreateUnit{
			ID:           uid.New(),
			OrgID:        c.OrgID,
			EquipmentID:  itemID,
			SerialNumber: serial.New(),
			IsActive:     true,
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Both types go through the same tx path; Bulk doesn't need atomicity but a
	// single code path is easier to maintain than two separate non-tx branches.
	var item *Equipment
	switch trackingType {
	case tracking.Bulk:
		item, err = s.repo.CreateBulk(ctx, tx, CreateBulkEquipment{
			ID:         itemID,
			BulkItemID: uid.New(),
			Base:       c.Base,
			TotalStock: c.TotalStock,
		})
	case tracking.Serialized:
		item, err = s.repo.CreateSerialized(ctx, tx, CreateSerializedEquipment{
			ID:    itemID,
			Base:  c.Base,
			Units: units,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Create: %w", err)
	}
	return item, nil
}

// ListParams holds the parameters for listing equipment.
type ListParams struct {
	OrgID    string
	Query    string
	Category string
	Filters  pagination.Filters
}

// List returns equipment for orgID filtered and paginated according to p.
// Returns all items when p.Filters.PageSize is math.MaxInt32.
func (s *Service) List(ctx context.Context, p ListParams) ([]Equipment, pagination.Metadata, error) {
	items, total, err := s.repo.List(ctx, p.OrgID, p.Query, p.Category, false, p.Filters)
	if err != nil {
		return nil, pagination.Metadata{}, fmt.Errorf("List: %w", err)
	}
	meta := pagination.CalculateMetadata(total, p.Filters.Page, p.Filters.PageSize)
	return items, meta, nil
}

// findByName returns the first equipment item in items whose name matches name,
// or nil when no match is found.
func findByName(items []Equipment, name string) *Equipment {
	for i := range items {
		if items[i].Name == name {
			return &items[i]
		}
	}
	return nil
}

// AssignContentByName resolves the member by name within the org then assigns it as content.
// Returns ErrNotFound when no equipment matches name.
func (s *Service) AssignContentByName(ctx context.Context, orgID, memberName string, a AssignContent) (*ContentItem, error) {
	all, _, err := s.repo.List(ctx, orgID, "", "", false, pagination.Filters{Page: 1, PageSize: math.MaxInt32})
	if err != nil {
		return nil, fmt.Errorf("AssignContentByName: %w", err)
	}
	member := findByName(all, memberName)
	if member == nil {
		return nil, fmt.Errorf("AssignContentByName: %w", ErrNotFound)
	}
	a.MemberID = member.ID
	item, err := s.AssignContent(ctx, a, *member)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// Get returns the inventory item with id.
func (s *Service) Get(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return item, nil
}

// UpdateDetails updates the details-tab fields. For bulk items it also updates
// the stock quantity; for serialized items only the inventory row is written.
func (s *Service) UpdateDetails(ctx context.Context, u UpdateEquipmentDetails) error {
	if u.Type == tracking.Bulk {
		if err := s.repo.UpdateDetailsBulk(ctx, u); err != nil {
			return fmt.Errorf("UpdateDetails: %w", err)
		}
		return nil
	}
	if err := s.repo.UpdateDetails(ctx, u); err != nil {
		return fmt.Errorf("UpdateDetails: %w", err)
	}
	return nil
}

// UpdatePricing updates the pricing-tab fields.
func (s *Service) UpdatePricing(ctx context.Context, u UpdateEquipmentPricing) error {
	if err := s.repo.UpdatePricing(ctx, u); err != nil {
		return fmt.Errorf("UpdatePricing: %w", err)
	}
	return nil
}

// UpdateProperties updates the properties-tab fields.
func (s *Service) UpdateProperties(ctx context.Context, u UpdateEquipmentProperties) error {
	if err := s.repo.UpdateProperties(ctx, u); err != nil {
		return fmt.Errorf("UpdateProperties: %w", err)
	}
	return nil
}

// SetImage links or unlinks a storage object from an inventory item.
func (s *Service) SetImage(ctx context.Context, si SetImage) error {
	if err := s.repo.SetImage(ctx, si); err != nil {
		return fmt.Errorf("SetImage: %w", err)
	}
	return nil
}

// Delete removes an inventory item.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}

// ListUnits returns all equipment items for the given inventory item, ordered by serial_number.
func (s *Service) ListUnits(ctx context.Context, inventoryID string) ([]Unit, error) {
	units, err := s.repo.ListUnits(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("ListUnits: %w", err)
	}
	return units, nil
}

// GetUnit returns the unit with id, or database.ErrNotFound when it does not exist.
func (s *Service) GetUnit(ctx context.Context, id string) (*Unit, error) {
	u, err := s.repo.GetUnit(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUnit: %w", err)
	}
	return u, nil
}

// AddUnit adds a new unit to the serialized inventory item.
// Returns ErrNotSerializedUnit when the equipment is not serialized.
func (s *Service) AddUnit(ctx context.Context, a AddUnit) (*Unit, error) {
	item, err := s.repo.Get(ctx, a.EquipmentID)
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	if item.TrackingType != tracking.Serialized {
		return nil, fmt.Errorf("AddUnit: %w", ErrNotSerializedUnit)
	}
	a.ID = uid.New()
	a.SerialNumber = serial.New()
	u, err := s.repo.AddUnit(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("AddUnit: %w", err)
	}
	return u, nil
}

// UpdateUnit updates a unit's editable fields.
func (s *Service) UpdateUnit(ctx context.Context, u UpdateUnit) error {
	if err := s.repo.UpdateUnit(ctx, u); err != nil {
		return fmt.Errorf("UpdateUnit: %w", err)
	}
	return nil
}

// BulkUpdateNextInspection sets the next inspection date for multiple units.
func (s *Service) BulkUpdateNextInspection(ctx context.Context, unitIDs []string, nextInspectionAt *int64) error {
	if err := s.repo.BulkUpdateNextInspection(ctx, unitIDs, nextInspectionAt); err != nil {
		return fmt.Errorf("BulkUpdateNextInspection: %w", err)
	}
	return nil
}

// DeleteUnit removes a unit from the serialized inventory item.
func (s *Service) DeleteUnit(ctx context.Context, unitID string) error {
	if err := s.repo.DeleteUnit(ctx, unitID); err != nil {
		return fmt.Errorf("DeleteUnit: %w", err)
	}
	return nil
}

// GetUnitsContainer fetches the equipment item and verifies it supports the units tab.
// Returns ErrNotFound when the item does not exist, ErrNoUnitsTab when it is bulk equipment.
func (s *Service) GetUnitsContainer(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUnitsContainer: %w", err)
	}
	if item.TrackingType != tracking.Serialized {
		return nil, fmt.Errorf("GetUnitsContainer: %w", ErrNoUnitsTab)
	}
	return item, nil
}

// GetContentContainer fetches the equipment item and verifies it supports the content tab.
// Returns ErrNotFound when the item does not exist, ErrNoContentTab when it is not a container.
func (s *Service) GetContentContainer(ctx context.Context, id string) (*Equipment, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetContentContainer: %w", err)
	}
	if item.Type != KitType {
		return nil, fmt.Errorf("GetContentContainer: %w", ErrNoContentTab)
	}
	return item, nil
}

// ListContent returns all content items for equipmentID.
func (s *Service) ListContent(ctx context.Context, equipmentID string) ([]ContentItem, error) {
	items, err := s.repo.ListContent(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("ListContent: %w", err)
	}
	return items, nil
}

// AssignContent adds an equipment item as content of a container.
// Returns ErrInvalidContent when the member is the container itself or is itself a container.
// Returns ErrConflict when the member is already assigned.
// Stock sufficiency is not enforced here; callers should surface warnings via ListContent.
func (s *Service) AssignContent(ctx context.Context, a AssignContent, member Equipment) (*ContentItem, error) {
	a.ID = uid.New()
	if member.ID == a.EquipmentID {
		return nil, fmt.Errorf("AssignContent: %w: an item cannot contain itself", ErrInvalidContent)
	}
	if member.Type == KitType {
		return nil, fmt.Errorf("AssignContent: %w: cannot add a container as content", ErrInvalidContent)
	}
	item, err := s.repo.AssignContent(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("AssignContent: %w", err)
	}
	return item, nil
}

// RemoveContent deletes the content entry with id.
func (s *Service) RemoveContent(ctx context.Context, id string) error {
	if err := s.repo.RemoveContent(ctx, id); err != nil {
		return fmt.Errorf("RemoveContent: %w", err)
	}
	return nil
}

// ListContainers returns all container equipment definitions that include memberID
// in their content definition.
func (s *Service) ListContainers(ctx context.Context, memberID string) ([]PartOf, error) {
	items, err := s.repo.ListContainersByMemberID(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("ListContainers: %w", err)
	}
	return items, nil
}
