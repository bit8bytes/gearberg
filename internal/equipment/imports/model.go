// Package imports handles CSV import staging and commit for inventory items.
package imports

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
	CurrentMa        string
	PowerMw          string
	Quantity         string
}

// ExpectedHeaders are the exact column headers the import CSV must have.
var ExpectedHeaders = []string{
	"Name", "Type", "Usage", "Category", "Manufacturer", "Location",
	"Rental Price", "Resale Price", "Notes",
	"Weight (g)", "Width (mm)", "Height (mm)", "Depth (mm)", "Voltage (V)", "Current (mA)", "Power (mW)",
	"Quantity",
}
