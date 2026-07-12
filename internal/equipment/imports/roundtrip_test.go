package imports_test

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/imports"
)

// buildCSV writes ExpectedHeaders followed by the rows produced by RowsForItem,
// mimicking exactly what the export handler does (minus the BOM and HTTP layer).
func buildCSV(item equipment.Equipment, mfrName string, units []equipment.Unit) string {
	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	_ = cw.Write(imports.ExpectedHeaders)
	for _, row := range imports.RowsForItem(item, mfrName, units) {
		_ = cw.Write(row)
	}
	cw.Flush()
	return buf.String()
}

func TestRoundTrip_bulk(t *testing.T) {
	stock := int64(7)
	item := equipment.Equipment{
		Name:         "Shure SM58",
		Type:         equipment.Bulk,
		UsageType:    equipment.Rental,
		CategoryName: "Audio",
		LocationName: "Main Warehouse",
		TotalStock:   stock,
		Notes:        "Cardioid dynamic mic",
		Pricing: equipment.Pricing{
			RentalPrice:   equipment.ParseCents("15.00"),
			PurchasePrice: equipment.ParseCents("99.00"),
		},
		Properties: equipment.Properties{
			Weight: equipment.ParseGrams("0.298"),
			Width:  equipment.ParseMillimeters("4.7"),
			Height: equipment.ParseMillimeters("4.7"),
			Depth:  equipment.ParseMillimeters("16.2"),
		},
	}

	rows, err := imports.ParseCSV(strings.NewReader(buildCSV(item, "Shure", nil)))
	if err != nil {
		t.Fatalf("ParseCSV: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	got := rows[0]

	check := func(field, want, got string) {
		t.Helper()
		if got != want {
			t.Errorf("%s: want %q, got %q", field, want, got)
		}
	}
	check("Name", "Shure SM58", got.Name)
	check("TypeLabel", "Bulk", got.TypeLabel)
	check("UsageTypeLabel", "Rental", got.UsageTypeLabel)
	check("CategoryName", "Audio", got.CategoryName)
	check("ManufacturerName", "Shure", got.ManufacturerName)
	check("LocationName", "Main Warehouse", got.LocationName)
	check("RentalPrice", "15.00", got.RentalPrice)
	check("ResalePrice", "99.00", got.ResalePrice)
	check("Notes", "Cardioid dynamic mic", got.Notes)
	check("Quantity", "7", got.Quantity)
	check("HasContent", "", got.HasContent)
	check("UnitSerialNumber", "", got.UnitSerialNumber)
}

func TestRoundTrip_serialized(t *testing.T) {
	purchasedAt := int64(1710460800) // 2024-03-15 UTC
	inspectAt := int64(1742083200)   // 2025-03-16 UTC

	item := equipment.Equipment{
		Name:         "Sony A7 IV",
		Type:         equipment.Serialized,
		UsageType:    equipment.Rental,
		CategoryName: "Camera",
		LocationName: "Main Warehouse",
		HasContent:   false,
		Notes:        "Full-frame mirrorless camera",
		Pricing: equipment.Pricing{
			RentalPrice:   equipment.ParseCents("80.00"),
			PurchasePrice: equipment.ParseCents("2800.00"),
		},
		Properties: equipment.Properties{
			Weight: equipment.ParseGrams("0.659"),
			Width:  equipment.ParseMillimeters("13.1"),
			Height: equipment.ParseMillimeters("9.6"),
			Depth:  equipment.ParseMillimeters("8.0"),
		},
	}
	units := []equipment.Unit{
		{
			StatusID:                 1,
			SerialNumber:             "SN-A7IV-001",
			ManufacturerSerialNumber: "7-000001",
			PurchasePrice:            equipment.ParseCents("2800.00"),
			PurchasedAt:              &purchasedAt,
			NextInspectionAt:         &inspectAt,
		},
		{
			StatusID:                 1,
			SerialNumber:             "SN-A7IV-002",
			ManufacturerSerialNumber: "7-000002",
			PurchasePrice:            equipment.ParseCents("2800.00"),
			PurchasedAt:              &purchasedAt,
			NextInspectionAt:         &inspectAt,
			Remark:                   "Minor scratch on top plate",
		},
	}

	rows, err := imports.ParseCSV(strings.NewReader(buildCSV(item, "Sony", units)))
	if err != nil {
		t.Fatalf("ParseCSV: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	check := func(field, want, got string) {
		t.Helper()
		if got != want {
			t.Errorf("%s: want %q, got %q", field, want, got)
		}
	}

	for i, got := range rows {
		check("Name", "Sony A7 IV", got.Name)
		check("TypeLabel", "Serialized", got.TypeLabel)
		check("UsageTypeLabel", "Rental", got.UsageTypeLabel)
		check("Quantity", "", got.Quantity)
		check("HasContent", "", got.HasContent)
		check("UnitIsActive", "TRUE", got.UnitIsActive)
		check("UnitSerialNumber", units[i].SerialNumber, got.UnitSerialNumber)
		check("UnitManufacturerSerial", units[i].ManufacturerSerialNumber, got.UnitManufacturerSerial)
		check("UnitPurchasePrice", "2800.00", got.UnitPurchasePrice)
		check("UnitRemark", units[i].Remark, got.UnitRemark)
	}
	check("UnitRemark", "Minor scratch on top plate", rows[1].UnitRemark)
}

func TestRoundTrip_serializedHasContent(t *testing.T) {
	item := equipment.Equipment{
		Name:         "Pelican 1510 Case",
		Type:         equipment.Serialized,
		UsageType:    equipment.Rental,
		CategoryName: "Case",
		LocationName: "Main Warehouse",
		HasContent:   true,
		Pricing: equipment.Pricing{
			RentalPrice:   equipment.ParseCents("10.00"),
			PurchasePrice: equipment.ParseCents("120.00"),
		},
	}
	purchasedAt := int64(1704844800) // 2024-01-10 UTC
	units := []equipment.Unit{
		{StatusID: 1, SerialNumber: "PC-1510-001", PurchasePrice: equipment.ParseCents("120.00"), PurchasedAt: &purchasedAt},
	}

	rows, err := imports.ParseCSV(strings.NewReader(buildCSV(item, "Pelican", units)))
	if err != nil {
		t.Fatalf("ParseCSV: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].HasContent != "TRUE" {
		t.Errorf("HasContent: want %q, got %q", "TRUE", rows[0].HasContent)
	}
	if rows[0].UnitSerialNumber != "PC-1510-001" {
		t.Errorf("UnitSerialNumber: want %q, got %q", "PC-1510-001", rows[0].UnitSerialNumber)
	}
}
