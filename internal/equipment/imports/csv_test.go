package imports_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bit8bytes/gearberg/internal/equipment/imports"
)

// TestParseCSV_roundtrip verifies that every column in ExpectedHeaders maps to
// the correct RawRow field. If ExpectedHeaders is reordered or readRows is not
// updated in step, this test breaks before any real data is affected.
func TestParseCSV_roundtrip(t *testing.T) {
	want := imports.RawRow{
		Name:                   "Shure SM58",
		TypeLabel:              "Bulk",
		UsageTypeLabel:         "Rental",
		CategoryName:           "Audio",
		ManufacturerName:       "Shure",
		LocationName:           "Main Warehouse",
		RentalPrice:            "15.00",
		ResalePrice:            "99.00",
		Notes:                  "Cardioid dynamic mic",
		WeightG:                "0.298",
		WidthMm:                "4.7",
		HeightMm:               "4.8",
		DepthMm:                "16.2",
		VoltageV:               "5",
		CurrentA:               "2.4",
		PowerW:                 "12",
		WireGaugeMM2X100:       "150",
		Quantity:               "7",
		HasContent:             "",
		UnitSerialNumber:       "",
		UnitManufacturerSerial: "",
		UnitPurchasePrice:      "",
		UnitPurchasedAt:        "",
		NextInspectionAt:       "",
		UnitIsActive:           "",
		UnitRemark:             "",
	}

	header := strings.Join(imports.ExpectedHeaders, ",")
	dataRow := strings.Join([]string{
		want.Name,
		want.TypeLabel,
		want.UsageTypeLabel,
		want.CategoryName,
		want.ManufacturerName,
		want.LocationName,
		want.RentalPrice,
		want.ResalePrice,
		want.Notes,
		want.WeightG,
		want.WidthMm,
		want.HeightMm,
		want.DepthMm,
		want.VoltageV,
		want.CurrentA,
		want.PowerW,
		want.WireGaugeMM2X100,
		want.Quantity,
		want.HasContent,
		want.UnitSerialNumber,
		want.UnitManufacturerSerial,
		want.UnitPurchasePrice,
		want.UnitPurchasedAt,
		want.NextInspectionAt,
		want.UnitIsActive,
		want.UnitRemark,
	}, ",")
	csv := header + "\n" + dataRow + "\n"

	rows, err := imports.ParseCSV(strings.NewReader(csv))
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
			t.Errorf("field %s: want %q, got %q", field, want, got)
		}
	}
	check("Name", want.Name, got.Name)
	check("TypeLabel", want.TypeLabel, got.TypeLabel)
	check("UsageTypeLabel", want.UsageTypeLabel, got.UsageTypeLabel)
	check("CategoryName", want.CategoryName, got.CategoryName)
	check("ManufacturerName", want.ManufacturerName, got.ManufacturerName)
	check("LocationName", want.LocationName, got.LocationName)
	check("RentalPrice", want.RentalPrice, got.RentalPrice)
	check("ResalePrice", want.ResalePrice, got.ResalePrice)
	check("Notes", want.Notes, got.Notes)
	check("WeightG", want.WeightG, got.WeightG)
	check("WidthMm", want.WidthMm, got.WidthMm)
	check("HeightMm", want.HeightMm, got.HeightMm)
	check("DepthMm", want.DepthMm, got.DepthMm)
	check("VoltageV", want.VoltageV, got.VoltageV)
	check("CurrentA", want.CurrentA, got.CurrentA)
	check("PowerW", want.PowerW, got.PowerW)
	check("WireGaugeMM2X100", want.WireGaugeMM2X100, got.WireGaugeMM2X100)
	check("Quantity", want.Quantity, got.Quantity)
	check("HasContent", want.HasContent, got.HasContent)
	check("UnitSerialNumber", want.UnitSerialNumber, got.UnitSerialNumber)
	check("UnitManufacturerSerial", want.UnitManufacturerSerial, got.UnitManufacturerSerial)
	check("UnitPurchasePrice", want.UnitPurchasePrice, got.UnitPurchasePrice)
	check("UnitPurchasedAt", want.UnitPurchasedAt, got.UnitPurchasedAt)
	check("NextInspectionAt", want.NextInspectionAt, got.NextInspectionAt)
	check("UnitIsActive", want.UnitIsActive, got.UnitIsActive)
	check("UnitRemark", want.UnitRemark, got.UnitRemark)
}

func TestParseCSV_wrongColumnCount(t *testing.T) {
	csv := "Name,Type\nFoo,Bar\n"
	_, err := imports.ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for wrong column count, got nil")
	}
}

func TestParseCSV_wrongColumnName(t *testing.T) {
	// Replace the first header with something unexpected.
	headers := make([]string, len(imports.ExpectedHeaders))
	copy(headers, imports.ExpectedHeaders)
	headers[0] = "ItemName" // was "Name"
	csv := strings.Join(headers, ",") + "\nFoo,Bulk,Rental,Audio,Shure,WH,15,99,,,,,,,,,,1\n"
	_, err := imports.ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for wrong column name, got nil")
	}
}

func TestParseCSV_noDataRows(t *testing.T) {
	csv := strings.Join(imports.ExpectedHeaders, ",") + "\n"
	_, err := imports.ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
}

// TestParseCSV_templateValid ensures TemplateCSV parses without error, so a
// change to ExpectedHeaders that is not reflected in template.csv breaks the
// build immediately rather than at runtime.
func TestParseCSV_templateValid(t *testing.T) {
	rows, err := imports.ParseCSV(bytes.NewReader(imports.TemplateCSV))
	if err != nil {
		t.Fatalf("TemplateCSV is not valid: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("TemplateCSV has no data rows")
	}
}

func TestParseCSV_stripsUTF8BOM(t *testing.T) {
	header := strings.Join(imports.ExpectedHeaders, ",")
	csv := "\xEF\xBB\xBF" + header + "\nShure SM58,Bulk,Rental,Audio,Shure,WH,15,99,,,,,,,,,,1,,,,,,,,\n"
	rows, err := imports.ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV with BOM: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}
