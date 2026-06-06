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
	ID               string
	ImportID         string
	OrgID            string
	RowNumber        int64
	Name             string
	TypeLabel        string
	UsageTypeLabel   string
	CategoryName     string
	ManufacturerName string
	TotalStock       string
	PurchasePrice    string
	RentalPrice      string
	Notes            string
	Status           string
	ErrorMessage     string
	Action           string
	ExistingItemID   *string
	CreatedAt        int64
}

// RawRow holds a parsed CSV data row before validation.
type RawRow struct {
	Name             string
	TypeLabel        string
	UsageTypeLabel   string
	CategoryName     string
	ManufacturerName string
	TotalStock       string
	PurchasePrice    string
	RentalPrice      string
	Notes            string
}

// ExpectedHeaders are the exact column headers the import CSV must have.
// They match the export format (including read-only columns Code, Created At, Updated At
// which are accepted but ignored on import).
var ExpectedHeaders = []string{
	"Code", "Name", "Type", "Usage", "Category", "Manufacturer",
	"Total Stock", "Purchase Price", "Rental Price", "Notes",
	"Created At", "Updated At",
}
