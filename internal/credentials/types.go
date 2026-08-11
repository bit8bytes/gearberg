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
package credentials

// Type represents a credential type. The integer value matches the id stored
// in the credential_types table, which is seeded at startup.
type Type int64

// Credential type IDs matching rows seeded in the credential_types table.
const (
	PasswordType Type = 1
)

// ID returns the database id for the credential type.
func (t Type) ID() int64 { return int64(t) }

// String returns the name stored in the credential_types table.
func (t Type) String() string {
	switch t {
	case PasswordType:
		return "password"
	default:
		return ""
	}
}
