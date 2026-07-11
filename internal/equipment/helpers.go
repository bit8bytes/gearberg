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
		return 1
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

// ParseCents parses a decimal price string and returns the value in cents. Returns nil when blank.
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

// ParseGrams parses a kg decimal string and returns the value in grams. Returns nil when blank.
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

// ParseMillimeters parses a cm decimal string and returns the value in mm. Returns nil when blank.
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

// ParseMilliwatts parses a W decimal string and returns the value in mW. Returns nil when blank.
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

// ParseMilliamps parses an A decimal string and returns the value in mA. Returns nil when blank.
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

// ParseVolts parses a decimal volt string and returns the value in millivolts. Returns nil when blank.
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

// ParseWireGauge parses a whole-number mm²×100 string. Returns nil when blank.
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

// parseUnixDate parses a YYYY-MM-DD date string and returns a Unix timestamp pointer. Returns nil when blank.
func parseUnixDate(s string) *int64 {
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
