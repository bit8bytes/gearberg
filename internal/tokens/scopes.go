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
package tokens

// Scope represents a token scope.
// The integer value matches the id stored in the token_scopes table,
// which is seeded at startup rather than via migrations.
type Scope int64

const (
	// PasswordReset is the scope for password reset tokens.
	PasswordReset Scope = 1
	// EmailVerification is the scope for email verification tokens.
	EmailVerification Scope = 2
)

// ID returns the database id for the token scope.
func (s Scope) ID() int64 { return int64(s) }

// String returns the name stored in the token_scopes table.
func (s Scope) String() string {
	switch s {
	case PasswordReset:
		return "password-reset"
	case EmailVerification:
		return "email-verification"
	default:
		return ""
	}
}
