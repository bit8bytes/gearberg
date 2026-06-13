// Package units provides form parsing, validation, and status types for serialized inventory units.
package units

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bit8bytes/toolbox/validator"
)

// Form holds the parsed form input and validation state for adding or editing a unit.
type Form struct {
	StatusID     string
	SerialNumber string
	Notes        string
	PurchasedAt  string // YYYY-MM-DD
	validator.Validator
}

// Parse reads unit form fields from r.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseForm(); err != nil {
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	return Form{
		StatusID:     strings.TrimSpace(r.PostForm.Get("status_id")),
		SerialNumber: strings.TrimSpace(r.PostForm.Get("serial_number")),
		Notes:        strings.TrimSpace(r.PostForm.Get("notes")),
		PurchasedAt:  strings.TrimSpace(r.PostForm.Get("purchased_at")),
	}, nil
}

// Validate checks UnitForm fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	if validator.NotBlank(f.SerialNumber) {
		f.Check(validator.MaxChars(f.SerialNumber, 200), "serial_number", "Cannot exceed 200 characters")
	}
	if validator.NotBlank(f.Notes) {
		f.Check(validator.MaxChars(f.Notes, 1000), "notes", "Cannot exceed 1000 characters")
	}
	if validator.NotBlank(f.PurchasedAt) {
		t, err := time.Parse("2006-01-02", f.PurchasedAt)
		if t.After(time.Now()) {
			f.AddError("purchased_at", "Date cannot be in the future")
		}
		f.Check(err == nil, "purchased_at", "Must be a valid date (YYYY-MM-DD)")
	}
	return f.Valid()
}

// PurchasedAtUnix parses the YYYY-MM-DD date string and returns a Unix timestamp,
// or nil when the field is blank.
func (f *Form) PurchasedAtUnix() *int64 {
	if !validator.NotBlank(f.PurchasedAt) {
		return nil
	}
	t, err := time.Parse("2006-01-02", f.PurchasedAt)
	if err != nil {
		return nil
	}
	v := t.UTC().Unix()
	return &v
}

// StatusIDInt64 returns the parsed StatusID value, defaulting to 1 (available) when blank or invalid.
func (f *Form) StatusIDInt64() int64 {
	n, err := strconv.ParseInt(f.StatusID, 10, 64)
	if err != nil || n < 1 {
		return 1
	}
	return n
}
