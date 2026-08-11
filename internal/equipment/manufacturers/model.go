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

// Package manufacturers handles manufacturer routes, business logic, and data access.
package manufacturers

import "errors"

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

// Manufacturer represents a manufacturer belonging to an org.
type Manufacturer struct {
	ID        string
	OrgID     string
	Name      string
	UpdatedAt int64
	CreatedAt int64
}

// CreateManufacturer holds the data required to create a new manufacturer.
type CreateManufacturer struct {
	ID    string
	OrgID string
	Name  string
}

// UpdateManufacturer holds the data required to update a manufacturer.
type UpdateManufacturer struct {
	ID   string
	Name string
}
