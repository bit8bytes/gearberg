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

// Kind represents whether equipment is a physical item or a virtual combination.
// The integer value matches the id stored in the equipment_types table,
// which is seeded at startup rather than via migrations.
type Kind int64

// Equipment kind identifiers seeded into the equipment_types table.
const (
	Physical Kind = 1
)

// ID returns the database id for the equipment kind.
func (k Kind) ID() int64 { return int64(k) }

// String returns the name stored in the equipment_types table.
func (k Kind) String() string {
	switch k {
	case Physical:
		return "physical"
	default:
		return ""
	}
}

// KindFromString returns the Kind matching name, or 0 when unknown.
func KindFromString(name string) Kind {
	switch name {
	case "physical":
		return Physical
	default:
		return 0
	}
}

// Label returns the human-friendly label for the equipment kind.
func (k Kind) Label() string {
	switch k {
	case Physical:
		return "Physical"
	default:
		return ""
	}
}
