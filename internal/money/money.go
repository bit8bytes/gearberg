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
