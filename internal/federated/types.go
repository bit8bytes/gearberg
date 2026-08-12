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

// Package federated manages federated identity records linking external OIDC
// providers to internal accounts.
package federated

// Provider represents an external identity provider. The integer value matches
// the id stored in the provider_types table, which is seeded at startup.
type Provider int64

// Provider IDs matching rows seeded in the provider_types table.
const (
	AuthentikProvider Provider = 1
)

// ByName returns the Provider matching the given name and whether it was found.
func ByName(name string) (Provider, bool) {
	switch name {
	case "authentik":
		return AuthentikProvider, true
	default:
		return 0, false
	}
}

// ID returns the database id for the provider.
func (p Provider) ID() int64 { return int64(p) }

// String returns the name stored in the provider_types table.
func (p Provider) String() string {
	switch p {
	case AuthentikProvider:
		return "authentik"
	default:
		return ""
	}
}
