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

// Row is a staged import row.
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
	Name                string
	TypeLabel           string
	TrackingLabel       string
	UsageTypeLabel      string
	CategoryName        string
	ManufacturerName    string
	LocationName        string
	RentalPrice         string
	ResalePrice         string
	Notes               string
	WeightG             string
	WidthMm             string
	HeightMm            string
	DepthMm             string
	VoltageV            string
	CurrentMa           string
	PowerMw             string
	WireGaugeMM2X100    string
	Quantity            string
	PurchasePrice       string
	PurchasedAt         string
	ManufacturerSerial  string
	NextInspectionAt    string
	IsActive            string
	Remark              string
}

// RawRow holds a parsed CSV data row before validation.
type RawRow struct {
	Name             string
	TypeLabel        string
	UsageTypeLabel   string
	CategoryName     string
	ManufacturerName string
	LocationName     string
	RentalPrice      string
	ResalePrice      string
	Notes            string
	WeightG          string
	WidthMm          string
	HeightMm         string
	DepthMm          string
	VoltageV         string
	CurrentA         string
	PowerW           string
	WireGaugeMM2X100 string
	Quantity         string
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
	"Quantity",
}
