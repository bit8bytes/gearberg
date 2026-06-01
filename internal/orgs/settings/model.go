// Package settings handles org settings routes, business logic, and data access.
package settings

import (
	"math"
	"strconv"
)

// OrgSettings holds the configurable settings for a org.
type OrgSettings struct {
	ID       string
	OrgID    string
	Currency string
	VatRate  int64
	Timezone string
}

// VatRatePercent converts the stored basis-point value back to a human-readable
// percentage string suitable for pre-filling the settings form (e.g. 1900 → "19").
func (s *OrgSettings) VatRatePercent() string {
	v := float64(s.VatRate) / 100
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// UpsertOrgSettings holds the data required to create or update org settings.
type UpsertOrgSettings struct {
	ID       string
	OrgID    string
	Currency string
	VatRate  int64
	Timezone string
}
