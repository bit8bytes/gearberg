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

// Package categories handles equipment category routes, business logic, and data access.
package categories

import "errors"

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

// EquipmentCategory represents an equipment category belonging to a org.
type EquipmentCategory struct {
	ID        string
	OrgID     string
	Name      string
	UpdatedAt int64
	CreatedAt int64
}

// CreateEquipmentCategory holds the data required to create a new equipment category.
type CreateEquipmentCategory struct {
	ID    string
	OrgID string
	Name  string
}

// UpdateEquipmentCategory holds the data required to update an equipment category.
type UpdateEquipmentCategory struct {
	ID   string
	Name string
}
