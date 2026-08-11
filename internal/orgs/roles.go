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
package orgs

// Role represents an organization role (Owner, Admin, Member, Viewer).
// The integer value matches the id stored in the org_roles table,
// which is seeded at startup rather than via migrations.
type Role int64

// Role IDs matching rows seeded in the org_roles table.
const (
	OwnerRole  Role = 1
	AdminRole  Role = 2
	MemberRole Role = 3
	ViewerRole Role = 4
)

// ID returns the database id for the role.
func (r Role) ID() int64 { return int64(r) }

// String returns the name stored in the org_roles table.
func (r Role) String() string {
	switch r {
	case OwnerRole:
		return "Owner"
	case AdminRole:
		return "Admin"
	case MemberRole:
		return "Member"
	case ViewerRole:
		return "Viewer"
	default:
		return ""
	}
}

// Rank returns the numeric rank for the role used in permission comparisons.
func (r Role) Rank() int64 {
	switch r {
	case OwnerRole:
		return 100
	case AdminRole:
		return 50
	case MemberRole:
		return 25
	case ViewerRole:
		return 0
	default:
		return 0
	}
}
