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

// Package uid wraps the application's ID generation and parsing.
package uid

import (
	"fmt"

	"github.com/segmentio/ksuid"
)

// Parse validates s as a well-formed ID and returns the canonical string form.
func Parse(s string) (string, error) {
	id, err := ksuid.Parse(s)
	if err != nil {
		return "", fmt.Errorf("uid.Parse: %w", err)
	}
	return id.String(), nil
}

// New returns a new unique ID.
func New() string {
	return ksuid.New().String()
}
