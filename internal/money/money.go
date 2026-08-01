// Package money provides monetary domain types: Currency (ISO-4217) and VatRate (basis points).
package money

import (
	"math"
	"strconv"
)

// ISO4217 lists the ISO-4217 codes exposed in the UI.
var ISO4217 = []string{
	"USD", "EUR", "GBP", "CHF", "CAD", "AUD", "JPY", "CNY", "INR", "BRL",
	"MXN", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF", "RON", "SGD", "HKD",
	"NZD", "ZAR", "TRY", "AED", "SAR",
}

// Currency is an ISO-4217 currency code (e.g. "EUR", "USD").
type Currency string

// String returns the currency code as a plain string.
func (c Currency) String() string { return string(c) }

// VatRate is a VAT rate stored as basis points (e.g. 1900 = 19%).
type VatRate int64

// Percent converts the stored basis-point value back to a human-readable
// percentage string suitable for pre-filling the settings form (e.g. 1900 → "19").
func (v VatRate) Percent() string {
	f := float64(v) / 100
	if f == math.Trunc(f) {
		return strconv.FormatFloat(f, 'f', 0, 64)
	}
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// Display formats the VAT rate as a percentage string for display (e.g. 1900 → "19%").
// Returns "" when zero.
func (v VatRate) Display() string {
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(float64(v)/100, 'f', -1, 64) + " %"
}

// Round converts a percentage float (e.g. 19.0) to a VatRate in basis points.
func Round(pct float64) VatRate {
	return VatRate(math.Round(pct * 100))
}
