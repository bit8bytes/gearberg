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

// Package imports provides imports functionality.
package imports

import (
	"strconv"
	"time"

	"github.com/bit8bytes/gearberg/internal/equipment"
)

// RowsForItem returns the CSV data rows for one equipment item using the column
// order defined by ExpectedHeaders.
// Bulk items produce a single row; serialized items produce one row per unit.
func RowsForItem(item equipment.Equipment, mfrName string, units []equipment.Unit) [][]string {
	hasContent := ""
	if item.HasContent {
		hasContent = "TRUE"
	}

	base := []string{
		item.Name,
		item.TrackingType.Label(),
		item.UsageType.Label(),
		item.CategoryName,
		mfrName,
		item.LocationName,
		item.Pricing.RentalPrice.ToDecimal(),
		item.Pricing.PurchasePrice.ToDecimal(),
		item.Notes,
		item.Properties.Weight.ToKG(),
		item.Properties.Width.ToCM(),
		item.Properties.Height.ToCM(),
		item.Properties.Depth.ToCM(),
		item.Properties.Voltage.ToV(),
		item.Properties.Current.ToA(),
		item.Properties.Power.ToW(),
		item.Properties.WireGauge.String(),
	}

	if item.TrackingType != equipment.Serialized {
		row := make([]string, len(ExpectedHeaders))
		copy(row, base)
		row[17] = strconv.FormatInt(item.TotalStock, 10)
		row[18] = hasContent
		return [][]string{row}
	}

	rows := make([][]string, 0, len(units))
	for _, u := range units {
		row := make([]string, len(ExpectedHeaders))
		copy(row, base)
		row[18] = hasContent
		row[19] = u.SerialNumber
		row[20] = u.ManufacturerSerialNumber
		row[21] = u.PurchasePrice.ToDecimal()
		row[22] = FormatExportDate(u.PurchasedAt)
		row[23] = FormatExportDate(u.NextInspectionAt)
		row[24] = FormatExportActive(u.IsActive())
		row[25] = u.Remark
		rows = append(rows, row)
	}
	return rows
}

// FormatExportDate formats a Unix timestamp pointer as YYYY-MM-DD, or "" if nil.
func FormatExportDate(ts *int64) string {
	if ts == nil {
		return ""
	}
	return time.Unix(*ts, 0).UTC().Format("2006-01-02")
}

// FormatExportActive returns "TRUE" for active, "FALSE" for inactive.
func FormatExportActive(active bool) string {
	if active {
		return "TRUE"
	}
	return "FALSE"
}
