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

// Package imports handles CSV import staging and commit for inventory items.
package imports

import _ "embed"

// Status values for a staged import row.
const (
	StatusNew   = "new"
	StatusError = "error"
)

// Action values for a staged import row.
const (
	ActionCreate = "create"
	ActionSkip   = "skip"
)

// Row is a staged import row persisted in equipment_imports.
type Row struct {
	ID                  string
	ImportID            string
	OrgID               string
	RowNumber           int64
	Status              string
	ErrorMessage        string
	Action              string
	ExistingEquipmentID *string
	ExistingItemID      *string
	CreatedAt           int64
	// Equipment fields
	Name               string
	TypeLabel          string
	TrackingLabel      string
	UsageTypeLabel     string
	CategoryName       string
	ManufacturerName   string
	LocationName       string
	PurchasePrice      string // equipment-level purchase price (reserved; not yet in CSV)
	RentalPrice        string
	ResalePrice        string
	Notes              string
	WeightG            string
	WidthMm            string
	HeightMm           string
	DepthMm            string
	VoltageMv          string
	CurrentMa          string
	PowerMw            string
	WireGaugeMM2X100   string
	Quantity           string
	EquipmentTypeLabel string
	// Unit fields (serialized items only)
	UnitSerialNumber       string
	UnitManufacturerSerial string
	UnitPurchasePrice      string
	UnitPurchasedAt        string
	NextInspectionAt       string
	UnitIsActive           string
	UnitRemark             string
}

// RawRow holds a parsed CSV data row before validation and staging.
type RawRow struct {
	// Equipment fields
	Name               string
	TypeLabel          string
	UsageTypeLabel     string
	CategoryName       string
	ManufacturerName   string
	LocationName       string
	RentalPrice        string
	ResalePrice        string
	Notes              string
	WeightG            string
	WidthMm            string
	HeightMm           string
	DepthMm            string
	VoltageV           string
	CurrentA           string
	PowerW             string
	WireGaugeMM2X100   string
	Quantity           string
	EquipmentTypeLabel string
	// Unit fields (serialized items only; blank for bulk)
	UnitSerialNumber       string
	UnitManufacturerSerial string
	UnitPurchasePrice      string
	UnitPurchasedAt        string
	NextInspectionAt       string
	UnitIsActive           string
	UnitRemark             string
}

// TemplateCSV is the pre-filled example CSV file served to users as a download template.
//
//go:embed template.csv
var TemplateCSV []byte

// ExpectedHeaders are the exact column headers the import CSV must have.
var ExpectedHeaders = []string{
	"Name", "Type", "Usage", "Category", "Manufacturer", "Location",
	"Rental Price", "Resale Price", "Notes",
	"Weight (kg)", "Width (cm)", "Height (cm)", "Depth (cm)", "Voltage (V)", "Current (A)", "Power (W)",
	"Wire Gauge (mm² ×100)",
	"Quantity", "Equipment Type",
	"Unit Serial Number", "Unit Manufacturer Serial", "Unit Purchase Price",
	"Unit Purchased At", "Next Inspection At", "Unit Active", "Unit Remark",
}
