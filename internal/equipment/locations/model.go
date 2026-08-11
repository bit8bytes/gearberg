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

// Package locations handles location routes, business logic, and data access.
package locations

import "errors"

var (
	// ErrNotFound is returned when a location does not exist.
	ErrNotFound = errors.New("location not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("location already exists")
	// ErrLimitExceeded is returned when an org's location limit is reached.
	ErrLimitExceeded = errors.New("location limit exceeded")
)

// Location represents a storage location belonging to an org.
type Location struct {
	ID                        string
	ParentWarehouseLocationID *string
	OrgID                     string
	Name                      string
}

// CreateLocation holds the data required to create a new location.
type CreateLocation struct {
	ID                        string
	ParentWarehouseLocationID *string
	OrgID                     string
	Name                      string
}

// UpdateLocation holds the data required to update a location.
type UpdateLocation struct {
	ID   string
	Name string
}
