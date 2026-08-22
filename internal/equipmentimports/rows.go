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

// Package equipmentimports provides imports functionality.
package equipmentimports

import (
	"strconv"
	"time"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
)

// headerIndex maps each ExpectedHeaders column name to its position.
// Built once at package init so RowsForItem can write by name, not by magic number.
var headerIndex = func() map[string]int {
	m := make(map[string]int, len(ExpectedHeaders))
	for i, name := range ExpectedHeaders {
		m[name] = i
	}
	return m
}()

func newRow(base []string) []string {
	row := make([]string, len(ExpectedHeaders))
	copy(row, base)
	return row
}

// RowsForItem returns the CSV data rows for one equipment item using the column
// order defined by ExpectedHeaders.
// Bulk items produce a single row; serialized items produce one row per unit.
func RowsForItem(item equipment.Equipment, mfrName string, units []equipment.Unit) [][]string {
	h := headerIndex
	base := make([]string, len(ExpectedHeaders))
	base[h["Name"]] = item.Name
	base[h["Type"]] = item.TrackingType.Label()
	base[h["Usage"]] = item.UsageType.Label()
	base[h["Category"]] = item.CategoryName
	base[h["Manufacturer"]] = mfrName
	base[h["Location"]] = item.LocationName
	base[h["Rental Price"]] = item.Pricing.RentalPrice.String()
	base[h["Resale Price"]] = item.Pricing.PurchasePrice.String()
	base[h["Notes"]] = item.Notes
	base[h["Weight (kg)"]] = item.Properties.Weight.String()
	base[h["Width (cm)"]] = item.Properties.Width.String()
	base[h["Height (cm)"]] = item.Properties.Height.String()
	base[h["Depth (cm)"]] = item.Properties.Depth.String()
	base[h["Voltage (V)"]] = item.Properties.Voltage.String()
	base[h["Current (A)"]] = item.Properties.Current.String()
	base[h["Power (W)"]] = item.Properties.Power.String()
	base[h["Wire Gauge (mm² ×100)"]] = item.Properties.WireGauge.String()
	base[h["Equipment Type"]] = item.Type.Label()

	if item.TrackingType != tracking.Serialized {
		row := newRow(base)
		row[h["Quantity"]] = strconv.FormatInt(item.TotalStock, 10)
		return [][]string{row}
	}

	rows := make([][]string, 0, len(units))
	for _, u := range units {
		row := newRow(base)
		row[h["Unit Serial Number"]] = u.SerialNumber
		row[h["Unit Manufacturer Serial"]] = u.ManufacturerSerialNumber
		row[h["Unit Purchase Price"]] = u.PurchasePrice.String()
		row[h["Unit Purchased At"]] = formatExportDate(u.PurchasedAt)
		row[h["Next Inspection At"]] = formatExportDate(u.NextInspectionAt)
		row[h["Unit Active"]] = formatExportActive(u.IsActive())
		row[h["Unit Remark"]] = u.Remark
		rows = append(rows, row)
	}
	return rows
}

// formatExportDate formats a Unix timestamp pointer as YYYY-MM-DD, or "" if nil.
func formatExportDate(ts *int64) string {
	if ts == nil {
		return ""
	}
	return time.Unix(*ts, 0).UTC().Format("2006-01-02")
}

// formatExportActive returns "TRUE" for active, "FALSE" for inactive.
func formatExportActive(active bool) string {
	if active {
		return "TRUE"
	}
	return "FALSE"
}
