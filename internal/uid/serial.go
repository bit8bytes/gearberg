// Package uid provides utilities for generating unique identifiers.
package uid

import (
	"crypto/rand"
	"math/big"
)

// serialAlphabet omits visually ambiguous characters (0, O, 1, I, l) so codes
// are safe to read aloud, type from a label, or scan from a QR print.
const serialAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// NewSerial returns an 8-character random string drawn from serialAlphabet.
// 32^8 ≈ 1 trillion combinations; suitable as a unique human-readable unit ID.
func NewSerial() string {
	b := make([]byte, 8)
	n := big.NewInt(int64(len(serialAlphabet)))
	for i := range b {
		idx, _ := rand.Int(rand.Reader, n)
		b[i] = serialAlphabet[idx.Int64()]
	}
	return string(b)
}
