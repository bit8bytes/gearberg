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

// Package equipmentimports handles CSV import staging and commit for inventory items.
package equipmentimports

import (
	_ "embed"
	"fmt"
	"strings"
)

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

// GroupedRow is a display-oriented view of staged rows collapsed by equipment name.
// Serialized rows that share a name are folded into one entry; Stock holds the
// unit count for serialized items and the quantity string for bulk.
type GroupedRow struct {
	RowNumber    int64
	Name         string
	TypeLabel    string
	CategoryName string
	Stock        string
	Status       string
	ErrorMessage string
}

// GroupRows collapses serialized staging rows that share a name into one GroupedRow
// and returns per-item counts for new and error items.
func GroupRows(staged []Row) (rows []GroupedRow, cntNew, cntError int) {
	type group struct {
		row    GroupedRow
		total  int
		hasErr bool
	}
	seen := make(map[string]*group)
	var order []string

	for _, r := range staged {
		if !strings.EqualFold(r.TypeLabel, "serialized") {
			rows = append(rows, GroupedRow{
				RowNumber:    r.RowNumber,
				Name:         r.Name,
				TypeLabel:    r.TypeLabel,
				CategoryName: r.CategoryName,
				Stock:        r.Quantity,
				Status:       r.Status,
				ErrorMessage: r.ErrorMessage,
			})
			if r.Status == StatusNew {
				cntNew++
			} else {
				cntError++
			}
			continue
		}

		key := strings.ToLower(r.Name)
		if _, ok := seen[key]; !ok {
			seen[key] = &group{row: GroupedRow{
				RowNumber:    r.RowNumber,
				Name:         r.Name,
				TypeLabel:    r.TypeLabel,
				CategoryName: r.CategoryName,
				Status:       StatusNew,
			}}
			order = append(order, key)
		}
		g := seen[key]
		g.total++
		if r.Status == StatusError && !g.hasErr {
			g.hasErr = true
			g.row.Status = StatusError
			g.row.ErrorMessage = r.ErrorMessage
		}
	}

	for _, key := range order {
		g := seen[key]
		g.row.Stock = fmt.Sprintf("%d", g.total)
		rows = append(rows, g.row)
		if g.hasErr {
			cntError++
		} else {
			cntNew++
		}
	}
	return
}

// Mapping links user-defined source column names to canonical Row field names.
// Key: header from the uploaded file; Value: Row field name.
type Mapping map[string]string

// RowState classifies a ProcessedRow for the UI and the commit gate.
type RowState string

const (
	StateValid   RowState = "valid"
	StateInvalid RowState = "invalid"
)

// ValidationError is a structured, field-level error message surfaced to the UI.
type ValidationError struct {
	Line   int    `json:"line"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// ProcessedRow wraps a Row with its validation state and collected errors.
// The pipeline operates on []ProcessedRow; Stage converts back to []Row for storage.
type ProcessedRow struct {
	State  RowState
	Errors []ValidationError
	Data   Row
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
