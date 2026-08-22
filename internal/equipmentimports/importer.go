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

package equipmentimports

import (
	"context"
	"fmt"
	"io"

	"github.com/bit8bytes/gearberg/internal/units"
	pkgcsv "github.com/bit8bytes/gearberg/pkg/csv"
)

// dbStr converts a *T (any int64-backed unit type) to its raw integer string,
// or "" if nil. Used to store DB-unit values in Row string fields.
func dbStr[T ~int64](v *T) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", int64(*v))
}

// Inspector reads only the header layer of a file to discover column names for
// the field-mapping UI step, without parsing the full dataset.
type Inspector interface {
	InspectHeaders(ctx context.Context, r io.Reader) ([]string, error)
}

// Reader parses an input source into records.
// Satisfied by *csv.Reader, or a future *json.Reader.
type Reader interface {
	Read(ctx context.Context, r io.Reader) ([]pkgcsv.Record, error)
}

// Step is a single pipeline stage applied to a batch of ProcessedRows.
// Steps are composable and run in order via RunValidation.
// A Step must never return early on a per-row error; instead it mutates
// the row's State to StateInvalid and appends to its Errors slice.
type Step func(ctx context.Context, rows []ProcessedRow) ([]ProcessedRow, error)

// Writer commits a processed batch to a destination.
// Close must be called after Write to release any held resources.
type Writer interface {
	Write(ctx context.Context, rows []Row) error
	Close() error
}

// RunValidation applies steps in order to rows, collecting validation errors
// per row without aborting early. Returns the annotated batch.
func RunValidation(ctx context.Context, rows []ProcessedRow, steps []Step) ([]ProcessedRow, error) {
	var err error
	for _, step := range steps {
		rows, err = step(ctx, rows)
		if err != nil {
			return nil, err
		}
	}
	return rows, nil
}

// MapRecords converts pkg/csv records into ProcessedRow values with all physical
// and monetary quantities already expressed in DB units (cents, grams, millimetres,
// millivolts, milliamps, milliwatts). Every row is initialised to StateValid;
// subsequent Steps mark rows invalid when constraints are violated.
// The caller supplies importID and orgID which are set on every row; IDs and
// status fields are left zero so Stage can assign them.
func MapRecords(records []pkgcsv.Record, importID, orgID string) []ProcessedRow {
	rows := make([]ProcessedRow, 0, len(records))
	for _, rec := range records {
		f := func(name string) string { return rec.Fields[name] }
		rows = append(rows, ProcessedRow{
			State: StateValid,
			Data: Row{
				ImportID:               importID,
				OrgID:                  orgID,
				Name:                   f("Name"),
				TypeLabel:              f("Type"),
				UsageTypeLabel:         f("Usage"),
				CategoryName:           f("Category"),
				ManufacturerName:       f("Manufacturer"),
				LocationName:           f("Location"),
				RentalPrice:            dbStr(units.ParseCents(f("Rental Price"))),
				ResalePrice:            dbStr(units.ParseCents(f("Resale Price"))),
				Notes:                  f("Notes"),
				WeightG:                dbStr(units.ParseGrams(f("Weight (kg)"))),
				WidthMm:                dbStr(units.ParseMillimeters(f("Width (cm)"))),
				HeightMm:               dbStr(units.ParseMillimeters(f("Height (cm)"))),
				DepthMm:                dbStr(units.ParseMillimeters(f("Depth (cm)"))),
				VoltageMv:              dbStr(units.ParseVolts(f("Voltage (V)"))),
				CurrentMa:              dbStr(units.ParseMilliamps(f("Current (A)"))),
				PowerMw:                dbStr(units.ParseMilliwatts(f("Power (W)"))),
				WireGaugeMM2X100:       dbStr(units.ParseWireGauge(f("Wire Gauge (mm² ×100)"))),
				Quantity:               f("Quantity"),
				EquipmentTypeLabel:     normalizeEquipmentTypeLabel(f("Equipment Type")),
				UnitSerialNumber:       f("Unit Serial Number"),
				UnitManufacturerSerial: f("Unit Manufacturer Serial"),
				UnitPurchasePrice:      dbStr(units.ParseCents(f("Unit Purchase Price"))),
				UnitPurchasedAt:        f("Unit Purchased At"),
				NextInspectionAt:       f("Next Inspection At"),
				UnitIsActive:           f("Unit Active"),
				UnitRemark:             f("Unit Remark"),
			},
		})
	}
	return rows
}
