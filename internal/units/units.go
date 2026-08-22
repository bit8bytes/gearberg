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

// Package units provides domain unit types for physical and monetary quantities.
package units

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Cents is a monetary amount in the smallest currency unit (e.g. 1999 = €19.99).
type Cents int64

// String formats the amount as a decimal string for form inputs (e.g. 1999 → "19.99").
// Returns "" when nil.
func (c *Cents) String() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(*c)/100)
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

// Grams is a weight stored in grams; users enter and see kilograms.
type Grams int64

// String returns the weight in kg as a string for form inputs (e.g. 1500 → "1.5"). Returns "" when nil.
func (g *Grams) String() string {
	if g == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*g)/1000)
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

// Millimeters is a length stored in millimetres; users enter and see centimetres.
type Millimeters int64

// String returns the length in cm as a string for form inputs (e.g. 100 → "10"). Returns "" when nil.
func (m *Millimeters) String() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*m)/10)
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

// Milliwatts is power stored in milliwatts; users enter and see watts.
type Milliwatts int64

// String returns the power in W as a string for form inputs (e.g. 1500 → "1.5"). Returns "" when nil.
func (m *Milliwatts) String() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*m)/1000)
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

// Milliamps is current stored in milliamps; users enter and see amps.
type Milliamps int64

// String returns the current in A as a string for form inputs (e.g. 500 → "0.5"). Returns "" when nil.
func (m *Milliamps) String() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*m)/1000)
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

// Millivolts is voltage stored and displayed as whole volts.
type Millivolts int64

// String returns the voltage as a string for form inputs. Returns "" when nil.
func (v *Millivolts) String() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%g", float64(*v)/1000)
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

// WireGauge stores wire cross-section area as mm²×100 (e.g. 150 = 1.5 mm²); displayed as-is.
type WireGauge int64

// String returns the raw mm²×100 value as a string for form inputs. Returns "" when nil.
func (w *WireGauge) String() string {
	if w == nil {
		return ""
	}
	return fmt.Sprintf("%d", *w)
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
