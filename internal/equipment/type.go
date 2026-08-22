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

// Type represents the type of an equipment item.
// The integer value matches the id stored in the equipment_types table,
// which is seeded at startup rather than via migrations.
type Type int64

// Equipment type identifiers seeded into the equipment_types table.
const (
	StandardType Type = 1
	KitType      Type = 2
)

// ID returns the database id for the equipment type.
func (t Type) ID() int64 { return int64(t) }

// String returns the name stored in the equipment_types table.
func (t Type) String() string {
	switch t {
	case StandardType:
		return "standard"
	case KitType:
		return "kit"
	default:
		return ""
	}
}

// Parse returns the Type matching name, or 0 when unknown.
func Parse(name string) Type {
	switch name {
	case "standard":
		return StandardType
	case "kit":
		return KitType
	default:
		return 0
	}
}

// ParseOrDefault returns the Type matching name, or Standard when unknown.
func ParseOrDefault(name string) Type {
	if t := Parse(name); t != 0 {
		return t
	}
	return StandardType
}

// Label returns the human-friendly label for the equipment type.
func (t Type) Label() string {
	switch t {
	case StandardType:
		return "Standard"
	case KitType:
		return "Kit"
	default:
		return ""
	}
}
