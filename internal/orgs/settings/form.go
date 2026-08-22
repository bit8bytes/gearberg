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
	"strings"

	"github.com/bit8bytes/gearberg/internal/money"
	"github.com/bit8bytes/gearberg/internal/timezone"
	"github.com/bit8bytes/toolbox/validator"
)

// Form holds the parsed form input and validation state for org settings requests.
type Form struct {
	Currency money.Currency
	VatRate  money.VatRate
	Timezone timezone.Timezone
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
	f.Currency = money.Currency(strings.TrimSpace(r.PostForm.Get("currency")))
	f.VatRate = money.ParseVatRate(strings.TrimSpace(r.PostForm.Get("vat_rate")))
	f.Timezone = timezone.Timezone(strings.TrimSpace(r.PostForm.Get("timezone")))
	return f, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.PermittedValue(string(f.Currency), money.ISO4217...), "currency", "Must be a valid ISO-4217 currency code")
	f.Check(f.VatRate >= 0 && f.VatRate <= 10000, "vat_rate", "Must be between 0 and 100")
	f.Check(validator.PermittedValue(string(f.Timezone), timezone.IANA...), "timezone", "Must be a valid IANA timezone")
	return f.Valid()
}

// Form pre-populates a Form from the stored OrgSettings. Safe to call on a nil receiver.
func (s *OrgSettings) Form() Form {
	f := Form{}
	f.Errors = make(map[string]string)
	if s == nil {
		return f
	}
	f.Currency = s.Currency
	f.VatRate = s.VatRate
	f.Timezone = s.Timezone
	return f
}
