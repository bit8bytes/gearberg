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
