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

// Package serial provides utilities for generating unique identifiers.
package serial

import (
	"crypto/rand"
	"math/big"
)

// serialAlphabet omits visually ambiguous characters (0, O, 1, I, l) so codes
// are safe to read aloud, type from a label, or scan from a QR print.
const serialAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// New returns an 8-character random string drawn from serialAlphabet.
// 32^8 ≈ 1 trillion combinations; suitable as a unique human-readable unit ID.
func New() string {
	b := make([]byte, 8)
	n := big.NewInt(int64(len(serialAlphabet)))
	for i := range b {
		idx, _ := rand.Int(rand.Reader, n)
		b[i] = serialAlphabet[idx.Int64()]
	}
	return string(b)
}
