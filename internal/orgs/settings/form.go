// Package settings handles org settings routes, business logic, and data access.
package settings

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

// PermittedCurrencies lists the ISO-4217 codes exposed in the UI.
var PermittedCurrencies = []string{
	"USD", "EUR", "GBP", "CHF", "CAD", "AUD", "JPY", "CNY", "INR", "BRL",
	"MXN", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF", "RON", "SGD", "HKD",
	"NZD", "ZAR", "TRY", "AED", "SAR",
}

// PermittedTimezones lists the IANA timezone identifiers exposed in the UI.
var PermittedTimezones = []string{
	"Africa/Cairo", "Africa/Johannesburg", "Africa/Lagos", "Africa/Nairobi",
	"America/Anchorage", "America/Argentina/Buenos_Aires", "America/Bogota",
	"America/Chicago", "America/Denver", "America/Los_Angeles", "America/Mexico_City",
	"America/New_York", "America/Phoenix", "America/Sao_Paulo", "America/Toronto",
	"America/Vancouver",
	"Asia/Bangkok", "Asia/Colombo", "Asia/Dubai", "Asia/Hong_Kong", "Asia/Jakarta",
	"Asia/Karachi", "Asia/Kolkata", "Asia/Riyadh", "Asia/Seoul", "Asia/Shanghai",
	"Asia/Singapore", "Asia/Tokyo",
	"Atlantic/Azores",
	"Australia/Adelaide", "Australia/Brisbane", "Australia/Melbourne",
	"Australia/Perth", "Australia/Sydney",
	"Europe/Amsterdam", "Europe/Athens", "Europe/Berlin", "Europe/Brussels",
	"Europe/Budapest", "Europe/Dublin", "Europe/Helsinki", "Europe/Istanbul",
	"Europe/Lisbon", "Europe/London", "Europe/Madrid", "Europe/Moscow",
	"Europe/Oslo", "Europe/Paris", "Europe/Prague", "Europe/Rome",
	"Europe/Stockholm", "Europe/Vienna", "Europe/Warsaw", "Europe/Zurich",
	"Pacific/Auckland", "Pacific/Honolulu",
	"UTC",
}

// Form holds the parsed form input and validation state for org settings requests.
type Form struct {
	Currency string
	VatRate  string // entered as a percentage e.g. "19" or "19,5"
	Timezone string
	validator.Validator
}

// Parse reads the org settings form fields from r.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseForm(); err != nil {
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	return Form{
		Currency: strings.TrimSpace(r.PostForm.Get("currency")),
		VatRate:  strings.TrimSpace(r.PostForm.Get("vat_rate")),
		Timezone: strings.TrimSpace(r.PostForm.Get("timezone")),
	}, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.PermittedValue(f.Currency, PermittedCurrencies...), "currency", "Must be a valid ISO-4217 currency code")
	if validator.NotBlank(f.VatRate) {
		v, err := strconv.ParseFloat(strings.ReplaceAll(f.VatRate, ",", "."), 64)
		f.Check(err == nil, "vat_rate", "Must be a valid number")
		f.Check(err != nil || (v >= 0 && v <= 100), "vat_rate", "Must be between 0 and 100")
	} else {
		f.AddError("vat_rate", "This field cannot be blank")
	}
	f.Check(validator.PermittedValue(f.Timezone, PermittedTimezones...), "timezone", "Must be a valid IANA timezone")
	return f.Valid()
}

// FormFromOrgSettings pre-populates a Form from stored OrgSettings.
func FormFromOrgSettings(s *OrgSettings) Form {
	if s == nil {
		return Form{}
	}
	return Form{
		Currency: s.Currency,
		VatRate:  s.VatRatePercent(),
		Timezone: s.Timezone,
	}
}

// VatRateBasisPoints returns the VAT rate as basis points (e.g. "19" → 1900).
// Call only after Validate() returns true.
func (f *Form) VatRateBasisPoints() int64 {
	v, _ := strconv.ParseFloat(strings.ReplaceAll(f.VatRate, ",", "."), 64)
	return int64(math.Round(v * 100))
}
