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
		hasContent = "true"
	}

	base := []string{
		item.Name,
		item.Type.Label(),
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

	if item.Type != equipment.Serialized {
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

// FormatExportActive returns "1" for active, "0" for inactive.
func FormatExportActive(active bool) string {
	if active {
		return "1"
	}
	return "0"
}
