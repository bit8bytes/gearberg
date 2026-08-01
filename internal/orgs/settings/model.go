// Package settings handles org settings routes, business logic, and data access.
package settings

import (
	"math"
	"strconv"
)

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

// Timezone is an IANA timezone identifier (e.g. "Europe/Berlin").
type Timezone string

// String returns the timezone identifier as a plain string.
func (t Timezone) String() string { return string(t) }

// OrgSettings holds the configurable settings for a org.
type OrgSettings struct {
	ID       string
	OrgID    string
	Currency Currency
	VatRate  VatRate
	Timezone Timezone
}

// UpdateOrgSettings holds the data required to update org settings.
type UpdateOrgSettings struct {
	OrgID    string
	Currency Currency
	VatRate  VatRate
	Timezone Timezone
}
