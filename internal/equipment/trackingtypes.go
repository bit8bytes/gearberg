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

// TrackingType represents an inventory type (Bulk or Serialized).
// The integer value matches the id stored in the inventory_types table,
// which is seeded at startup rather than via migrations.
type TrackingType int64

// Equipment type identifiers seeded into the inventory_types table.
const (
	Bulk       TrackingType = 1
	Serialized TrackingType = 2
)

// ID returns the database id for the inventory type.
func (k TrackingType) ID() int64 { return int64(k) }

// String returns the name stored in the inventory_types table.
func (k TrackingType) String() string {
	switch k {
	case Bulk:
		return "bulk"
	case Serialized:
		return "serialized"
	default:
		return ""
	}
}

// TrackingTypeFromString returns the TrackingType matching name ("bulk" or "serialized"), or 0 when unknown.
func TrackingTypeFromString(name string) TrackingType {
	switch name {
	case "bulk":
		return Bulk
	case "serialized":
		return Serialized
	default:
		return 0
	}
}

// Label returns the human-friendly label for the inventory type.
func (k TrackingType) Label() string {
	switch k {
	case Bulk:
		return "Bulk"
	case Serialized:
		return "Serialized"
	default:
		return ""
	}
}
