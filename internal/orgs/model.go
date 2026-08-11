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

import "errors"

var (
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("org already exists")
	// ErrLimitExceeded is returned when the org limit is reached.
	ErrLimitExceeded = errors.New("org limit exceeded")
)

// Org represents a org entity.
type Org struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}
