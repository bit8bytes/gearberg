package imports

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genimports "github.com/bit8bytes/gearberg/internal/database/queries/gen/invimports"
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
		ID:               row.ID,
		ImportID:         row.ImportID,
		OrgID:            row.OrgID,
		RowNumber:        row.RowNumber,
		Name:             row.Name,
		TypeLabel:        row.TypeLabel,
		UsageTypeLabel:   row.UsageTypeLabel,
		CategoryName:     row.CategoryName,
		ManufacturerName: row.ManufacturerName,
		TotalStock:       row.TotalStock,
		PurchasePrice:    row.PurchasePrice,
		RentalPrice:      row.RentalPrice,
		Notes:            row.Notes,
		Status:           row.Status,
		ErrorMessage:     row.ErrorMessage,
		Action:           row.Action,
		ExistingItemID:   database.NullString(row.ExistingItemID),
	})
	if err != nil {
		return nil, fmt.Errorf("Insert: %w", err)
	}
	out := fromRecord(rec)
	return &out, nil
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

func fromRecord(rec genimports.InventoryImport) Row {
	return Row{
		ID:               rec.ID,
		ImportID:         rec.ImportID,
		OrgID:            rec.OrgID,
		RowNumber:        rec.RowNumber,
		Name:             rec.Name,
		TypeLabel:        rec.TypeLabel,
		UsageTypeLabel:   rec.UsageTypeLabel,
		CategoryName:     rec.CategoryName,
		ManufacturerName: rec.ManufacturerName,
		TotalStock:       rec.TotalStock,
		PurchasePrice:    rec.PurchasePrice,
		RentalPrice:      rec.RentalPrice,
		Notes:            rec.Notes,
		Status:           rec.Status,
		ErrorMessage:     rec.ErrorMessage,
		Action:           rec.Action,
		ExistingItemID:   database.StringPtr(rec.ExistingItemID),
		CreatedAt:        rec.CreatedAt,
	}
}
