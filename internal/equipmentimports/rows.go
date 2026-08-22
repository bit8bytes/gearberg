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

// RowsForItem returns the CSV data rows for one equipment item using the column
// order defined by ExpectedHeaders.
// Bulk items produce a single row; serialized items produce one row per unit.
func RowsForItem(item equipment.Equipment, mfrName string, units []equipment.Unit) [][]string {
	equipmentTypeLabel := item.Type.Label()

	base := []string{
		item.Name,
		item.TrackingType.Label(),
		item.UsageType.Label(),
		item.CategoryName,
		mfrName,
		item.LocationName,
		item.Pricing.RentalPrice.String(),
		item.Pricing.PurchasePrice.String(),
		item.Notes,
		item.Properties.Weight.String(),
		item.Properties.Width.String(),
		item.Properties.Height.String(),
		item.Properties.Depth.String(),
		item.Properties.Voltage.String(),
		item.Properties.Current.String(),
		item.Properties.Power.String(),
		item.Properties.WireGauge.String(),
	}

	if item.TrackingType != tracking.Serialized {
		row := make([]string, len(ExpectedHeaders))
		copy(row, base)
		row[17] = strconv.FormatInt(item.TotalStock, 10)
		row[18] = equipmentTypeLabel
		return [][]string{row}
	}

	rows := make([][]string, 0, len(units))
	for _, u := range units {
		row := make([]string, len(ExpectedHeaders))
		copy(row, base)
		row[18] = equipmentTypeLabel
		row[19] = u.SerialNumber
		row[20] = u.ManufacturerSerialNumber
		row[21] = u.PurchasePrice.String()
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
