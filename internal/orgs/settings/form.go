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

// Package settings handles org settings routes, business logic, and data access.
package settings

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bit8bytes/gearberg/internal/money"
	"github.com/bit8bytes/gearberg/internal/timezone"
	"github.com/bit8bytes/toolbox/validator"
)

// Form holds the parsed form input and validation state for org settings requests.
type Form struct {
	Currency string
	VatRate  string // entered as a percentage e.g. "19" or "19,5"
	Timezone string
	validator.Validator
}

// NewForm returns a Form with an initialized Errors map, safe for template rendering.
func NewForm() *Form {
	f := &Form{}
	f.Errors = make(map[string]string)
	return f
}

// Parse reads the org settings form fields from r.
func Parse(r *http.Request) (Form, error) {
	f := Form{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.Currency = strings.TrimSpace(r.PostForm.Get("currency"))
	f.VatRate = strings.TrimSpace(r.PostForm.Get("vat_rate"))
	f.Timezone = strings.TrimSpace(r.PostForm.Get("timezone"))
	return f, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.PermittedValue(f.Currency, money.ISO4217...), "currency", "Must be a valid ISO-4217 currency code")
	if validator.NotBlank(f.VatRate) {
		v, err := strconv.ParseFloat(strings.ReplaceAll(f.VatRate, ",", "."), 64)
		f.Check(err == nil, "vat_rate", "Must be a valid number")
		f.Check(err != nil || (v >= 0 && v <= 100), "vat_rate", "Must be between 0 and 100")
	} else {
		f.AddError("vat_rate", "This field cannot be blank")
	}
	f.Check(validator.PermittedValue(f.Timezone, timezone.IANA...), "timezone", "Must be a valid IANA timezone")
	return f.Valid()
}

// FormFromOrgSettings pre-populates a Form from stored OrgSettings.
func FormFromOrgSettings(s *OrgSettings) Form {
	f := Form{}
	f.Errors = make(map[string]string)
	if s == nil {
		return f
	}
	f.Currency = s.Currency.String()
	f.VatRate = s.VatRate.Percent()
	f.Timezone = s.Timezone.String()
	return f
}

// ParsedVatRate converts the entered percentage string to a VatRate (basis points).
// Call only after Validate() returns true.
func (f *Form) ParsedVatRate() money.VatRate {
	v, _ := strconv.ParseFloat(strings.ReplaceAll(f.VatRate, ",", "."), 64)
	return money.Round(v)
}
