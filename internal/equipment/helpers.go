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

// Package equipment provides equipment functionality.
package equipment

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// parseCheckbox returns 1 if s is non-empty (checkbox was checked), 0 otherwise.
func parseCheckbox(s string) int64 {
	if strings.TrimSpace(s) != "" {
		return 1
	}
	return 0
}

// ParseQuantity parses a whole-number string. Returns 0 when blank or unparseable.
func ParseQuantity(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 {
		return 0
	}
	return n
}

// ParseCents parses a decimal price string and returns a pointer. Returns nil when blank or invalid.
func ParseCents(s string) *Cents {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := Cents(math.Round(f * 100))
	return &v
}

// ParseGrams parses a kg decimal string and returns a pointer. Returns nil when blank or invalid.
func ParseGrams(s string) *Grams {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := Grams(math.Round(f * 1000))
	return &v
}

// ParseMillimeters parses a cm decimal string and returns a pointer. Returns nil when blank or invalid.
func ParseMillimeters(s string) *Millimeters {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := Millimeters(math.Round(f * 10))
	return &v
}

// ParseMilliwatts parses a W decimal string and returns a pointer. Returns nil when blank or invalid.
func ParseMilliwatts(s string) *Milliwatts {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := Milliwatts(math.Round(f * 1000))
	return &v
}

// ParseMilliamps parses an A decimal string and returns a pointer. Returns nil when blank or invalid.
func ParseMilliamps(s string) *Milliamps {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := Milliamps(math.Round(f * 1000))
	return &v
}

// ParseVolts parses a V decimal string and returns a pointer. Returns nil when blank or invalid.
func ParseVolts(s string) *Millivolts {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	v := Millivolts(math.Round(f * 1000))
	return &v
}

// ParseWireGauge parses a whole-number mm²×100 string and returns a pointer. Returns nil when blank or invalid.
func ParseWireGauge(s string) *WireGauge {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	v := WireGauge(n)
	return &v
}

// ParseDate parses a YYYY-MM-DD date string and returns a Unix timestamp pointer. Returns nil when blank.
func ParseDate(s string) *int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	v := t.UTC().Unix()
	return &v
}
