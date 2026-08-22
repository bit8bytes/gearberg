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

// Package equipmentimports provides imports functionality.
package equipmentimports

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genimports "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipmentimports"
)

// Repository provides data access for import staging rows.
type Repository struct {
	q genimports.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{q: genimports.New(db)}
}

// Insert inserts a staged import row.
func (r *Repository) Insert(ctx context.Context, row Row) (*Row, error) {
	rec, err := r.q.InsertImportRow(ctx, genimports.InsertImportRowParams{
		ID:                     row.ID,
		ImportID:               row.ImportID,
		OrgID:                  row.OrgID,
		RowNumber:              row.RowNumber,
		Status:                 row.Status,
		ErrorMessage:           row.ErrorMessage,
		Action:                 row.Action,
		ExistingEquipmentID:    database.NullString(row.ExistingEquipmentID),
		ExistingItemID:         database.NullString(row.ExistingItemID),
		Name:                   row.Name,
		TypeLabel:              row.TypeLabel,
		TrackingLabel:          row.TrackingLabel,
		UsageTypeLabel:         row.UsageTypeLabel,
		CategoryName:           row.CategoryName,
		ManufacturerName:       row.ManufacturerName,
		LocationName:           row.LocationName,
		PurchasePrice:          row.PurchasePrice,
		RentalPrice:            row.RentalPrice,
		ResalePrice:            row.ResalePrice,
		Notes:                  row.Notes,
		WeightG:                row.WeightG,
		WidthMm:                row.WidthMm,
		HeightMm:               row.HeightMm,
		DepthMm:                row.DepthMm,
		VoltageMv:              row.VoltageMv,
		CurrentMa:              row.CurrentMa,
		PowerMw:                row.PowerMw,
		WireGaugeMm2X100:       row.WireGaugeMM2X100,
		Quantity:               row.Quantity,
		EquipmentTypeLabel:     row.EquipmentTypeLabel,
		UnitSerialNumber:       row.UnitSerialNumber,
		UnitManufacturerSerial: row.UnitManufacturerSerial,
		UnitPurchasePrice:      row.UnitPurchasePrice,
		UnitPurchasedAt:        row.UnitPurchasedAt,
		NextInspectionAt:       row.NextInspectionAt,
		UnitIsActive:           row.UnitIsActive,
		UnitRemark:             row.UnitRemark,
	})
	if err != nil {
		return nil, fmt.Errorf("Insert: %w", err)
	}
	out := fromRecord(rec)
	return &out, nil
}

// GetByID returns a single staged import row.
func (r *Repository) GetByID(ctx context.Context, id string) (*Row, error) {
	rec, err := r.q.GetImportRow(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	out := fromRecord(rec)
	return &out, nil
}

// UpdateAction updates the action field of a staged import row.
func (r *Repository) UpdateAction(ctx context.Context, id, action string) error {
	if err := r.q.UpdateImportRowAction(ctx, genimports.UpdateImportRowActionParams{
		ID:     id,
		Action: action,
	}); err != nil {
		return fmt.Errorf("UpdateAction: %w", err)
	}
	return nil
}

// DeleteByOrgID deletes all staging rows for the org.
func (r *Repository) DeleteByOrgID(ctx context.Context, orgID string) error {
	if err := r.q.DeleteImportsByOrgID(ctx, orgID); err != nil {
		return fmt.Errorf("DeleteByOrgID: %w", err)
	}
	return nil
}

// DeleteByImportID deletes all staging rows for a specific import.
func (r *Repository) DeleteByImportID(ctx context.Context, importID string) error {
	if err := r.q.DeleteImportsByImportID(ctx, importID); err != nil {
		return fmt.Errorf("DeleteByImportID: %w", err)
	}
	return nil
}

// ListByImportID returns all staged rows for an import, ordered by row_number.
func (r *Repository) ListByImportID(ctx context.Context, importID string) ([]Row, error) {
	recs, err := r.q.ListImportRowsByImportID(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("ListByImportID: %w", err)
	}
	rows := make([]Row, len(recs))
	for i, rec := range recs {
		rows[i] = fromRecord(rec)
	}
	return rows, nil
}

func fromRecord(rec genimports.EquipmentImport) Row {
	return Row{
		ID:                     rec.ID,
		ImportID:               rec.ImportID,
		OrgID:                  rec.OrgID,
		RowNumber:              rec.RowNumber,
		Status:                 rec.Status,
		ErrorMessage:           rec.ErrorMessage,
		Action:                 rec.Action,
		ExistingEquipmentID:    database.StringPtr(rec.ExistingEquipmentID),
		ExistingItemID:         database.StringPtr(rec.ExistingItemID),
		CreatedAt:              rec.CreatedAt,
		Name:                   rec.Name,
		TypeLabel:              rec.TypeLabel,
		TrackingLabel:          rec.TrackingLabel,
		UsageTypeLabel:         rec.UsageTypeLabel,
		CategoryName:           rec.CategoryName,
		ManufacturerName:       rec.ManufacturerName,
		LocationName:           rec.LocationName,
		PurchasePrice:          rec.PurchasePrice,
		RentalPrice:            rec.RentalPrice,
		ResalePrice:            rec.ResalePrice,
		Notes:                  rec.Notes,
		WeightG:                rec.WeightG,
		WidthMm:                rec.WidthMm,
		HeightMm:               rec.HeightMm,
		DepthMm:                rec.DepthMm,
		VoltageMv:              rec.VoltageMv,
		CurrentMa:              rec.CurrentMa,
		PowerMw:                rec.PowerMw,
		WireGaugeMM2X100:       rec.WireGaugeMm2X100,
		Quantity:               rec.Quantity,
		EquipmentTypeLabel:     rec.EquipmentTypeLabel,
		UnitSerialNumber:       rec.UnitSerialNumber,
		UnitManufacturerSerial: rec.UnitManufacturerSerial,
		UnitPurchasePrice:      rec.UnitPurchasePrice,
		UnitPurchasedAt:        rec.UnitPurchasedAt,
		NextInspectionAt:       rec.NextInspectionAt,
		UnitIsActive:           rec.UnitIsActive,
		UnitRemark:             rec.UnitRemark,
	}
}
