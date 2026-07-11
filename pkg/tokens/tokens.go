// Package tokens provides utilities for generating and hashing secure random tokens.
package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
)

// Token is 32 cryptographically random bytes.
type Token [32]byte

// Generate returns a new cryptographically secure random Token.
func Generate() Token {
	var token Token
	_, _ = io.ReadFull(rand.Reader, token[:])
	return token
}

// String returns the token base64 URL-encoded without padding.
func (t Token) String() string {
	return base64.RawURLEncoding.EncodeToString(t[:])
}

// Hex returns the token as a lowercase hexadecimal string.
func (t Token) Hex() string {
	return hex.EncodeToString(t[:])
}

// ShortToken is 8 cryptographically random bytes, suitable for trace IDs and
// other short-lived identifiers where a full 32-byte token is unnecessary.
type ShortToken [8]byte

// GenerateShort returns a new cryptographically secure random ShortToken.
func GenerateShort() ShortToken {
	var token ShortToken
	_, _ = io.ReadFull(rand.Reader, token[:])
	return token
}

// Hex returns the token as a lowercase hexadecimal string.
func (t ShortToken) Hex() string {
	return hex.EncodeToString(t[:])
}

// Hash returns the SHA-256 hash of a plaintext token string (e.g. read back from a cookie).
func Hash(plaintext string) []byte {
	sum := sha256.Sum256([]byte(plaintext))
	return sum[:]
}
