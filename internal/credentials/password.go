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
