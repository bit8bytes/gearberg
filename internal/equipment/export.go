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

package equipment

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
	"github.com/bit8bytes/gearberg/internal/equipment/usage"
)

// ExportRow is one CSV row matching the import template column order exactly.
type ExportRow struct {
	Name         string
	EquipmentType string // Standard or Kit
	TrackingType string // Bulk or Serialized
	Usage        string
	Category     string
	Manufacturer string
	Location     string
	RentalPrice  string
	ResalePrice  string
	Notes        string
	WeightKg     string
	WidthCm      string
	HeightCm     string
	DepthCm      string
	VoltageV     string
	CurrentA     string
	PowerW       string
	WireGauge    string
	Quantity     string
	// unit fields (serialized only)
	UnitSerialNumber       string
	UnitManufacturerSerial string
	UnitPurchasePrice      string
	UnitPurchasedAt        string
	NextInspectionAt       string
	UnitActive             string
	UnitRemark             string
}

var exportHeaders = []string{
	"Name", "Type", "Tracking Type", "Usage", "Category", "Manufacturer", "Location",
	"Rental Price", "Resale Price", "Notes",
	"Weight (kg)", "Width (cm)", "Height (cm)", "Depth (cm)",
	"Voltage (V)", "Current (A)", "Power (W)", "Wire Gauge (mm² ×100)",
	"Quantity",
	"Unit Serial Number", "Unit Manufacturer Serial", "Unit Purchase Price",
	"Unit Purchased At", "Next Inspection At", "Unit Active", "Unit Remark",
}

// WriteExportCSV writes rows to w in import-template column order.
func WriteExportCSV(w io.Writer, rows []ExportRow) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(exportHeaders); err != nil {
		return fmt.Errorf("WriteExportCSV: write header: %w", err)
	}
	for _, r := range rows {
		record := []string{
			r.Name, r.EquipmentType, r.TrackingType, r.Usage, r.Category, r.Manufacturer, r.Location,
			r.RentalPrice, r.ResalePrice, r.Notes,
			r.WeightKg, r.WidthCm, r.HeightCm, r.DepthCm,
			r.VoltageV, r.CurrentA, r.PowerW, r.WireGauge,
			r.Quantity,
			r.UnitSerialNumber, r.UnitManufacturerSerial, r.UnitPurchasePrice,
			r.UnitPurchasedAt, r.NextInspectionAt, r.UnitActive, r.UnitRemark,
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("WriteExportCSV: write row: %w", err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("WriteExportCSV: flush: %w", err)
	}
	return nil
}

// ExportCSV fetches all non-archived equipment for orgID and writes them as CSV to w.
func (s *Service) ExportCSV(ctx context.Context, w io.Writer, orgID string) error {
	rows, err := s.repo.Export(ctx, orgID)
	if err != nil {
		return fmt.Errorf("ExportCSV: %w", err)
	}
	return WriteExportCSV(w, rows)
}

func formatUnixDate(v *int64) string {
	if v == nil {
		return ""
	}
	return time.Unix(*v, 0).UTC().Format("2006-01-02")
}

func equipmentTypeLabel(name string) string {
	return ParseOrDefault(name).String()
}

func formatActive(isActive int64) string {
	if isActive == 1 {
		return "true"
	}
	return "false"
}

func usageLabel(name string) string {
	return usage.Parse(name).Label()
}

func trackingLabel(name string) string {
	return tracking.Parse(name).Label()
}

func formatQuantity(q int64) string {
	return strconv.FormatInt(q, 10)
}
