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

// Package credentials provides credentials functionality.
package credentials

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

// Password hashes and verifies passwords using argon2id with default parameters.
type Password struct{}

// CreateHash returns an encoded argon2id hash of plaintext.
func (h *Password) CreateHash(plaintext string) ([]byte, error) {
	s, err := argon2id.CreateHash(plaintext, argon2id.DefaultParams)
	return []byte(s), err
}

// ComparePasswordAndHash reports whether plaintext matches the encoded argon2id hash.
func (h *Password) ComparePasswordAndHash(plaintext string, hash []byte) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(plaintext, string(hash))
	if err != nil {
		return false, fmt.Errorf("credentials.Hasher.ComparePasswordAndHash: %w", err)
	}
	return match, nil
}
