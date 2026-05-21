// Package settings handles company settings routes, business logic, and data access.
package settings

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

// Form holds the parsed form input and validation state for company settings requests.
type Form struct {
	Currency string
	VatRate  float64
	Timezone string
	validator.Validator
}

// Parse reads the company settings form fields from r.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseForm(); err != nil {
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	vatRate, _ := strconv.ParseFloat(strings.TrimSpace(r.PostForm.Get("vat_rate")), 64)
	return Form{
		Currency: strings.TrimSpace(r.PostForm.Get("currency")),
		VatRate:  vatRate,
		Timezone: strings.TrimSpace(r.PostForm.Get("timezone")),
	}, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.NotBlank(f.Currency), "currency", "This field cannot be blank")
	f.Check(f.VatRate >= 0.00 && f.VatRate <= 1.00, "vat_rate", "Must be between 0.00 and 1.00")
	f.Check(validator.NotBlank(f.Timezone), "timezone", "This field cannot be blank")
	return f.Valid()
}
