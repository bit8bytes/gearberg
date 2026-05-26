// Package settings handles company settings routes, business logic, and data access.
package settings

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

// permittedCurrencies lists the ISO-4217 codes exposed in the UI.
var permittedCurrencies = []string{
	"USD", "EUR", "GBP", "CHF", "CAD", "AUD", "JPY", "CNY", "INR", "BRL",
	"MXN", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF", "RON", "SGD", "HKD",
	"NZD", "ZAR", "TRY", "AED", "SAR",
}

// permittedTimezones lists the IANA timezone identifiers exposed in the UI.
var permittedTimezones = []string{
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
	f.Check(validator.PermittedValue(f.Currency, permittedCurrencies...), "currency", "Must be a valid ISO-4217 currency code")
	f.Check(f.VatRate >= 0.00 && f.VatRate <= 1.00, "vat_rate", "Must be between 0.00 and 1.00")
	f.Check(validator.PermittedValue(f.Timezone, permittedTimezones...), "timezone", "Must be a valid IANA timezone")
	return f.Valid()
}
