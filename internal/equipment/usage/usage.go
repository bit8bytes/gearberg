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

// Package usage defines equipment usage classifications (rental, sale, etc.).
package usage

// Type represents how an inventory item is used (rental or sale).
// The integer value matches the id stored in the usage_types table,
// which is seeded at startup rather than via migrations.
type Type int64

// Usage type identifiers seeded into the usage_types table.
const (
	Rental Type = 1
)

// ID returns the database id for the usage type.
func (u Type) ID() int64 { return int64(u) }

// Parse returns the Usage matching name, or 0 when unknown.
func Parse(name string) Type {
	switch name {
	case "rental":
		return Rental
	default:
		return 0
	}
}

// Label returns the human-friendly label for the usage type.
func (u Type) Label() string {
	switch u {
	case Rental:
		return "Rental"
	default:
		return ""
	}
}

// String returns the name stored in the usage_types table.
func (u Type) String() string {
	switch u {
	case Rental:
		return "rental"
	default:
		return ""
	}
}
