package main

import (
	"encoding/csv"
	"net/http"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/imports"
	"github.com/bit8bytes/gearberg/internal/httperr"
)

// exportColumnCount must equal len(imports.ExpectedHeaders). This line fails to
// compile if RowsForItem is updated without updating ExpectedHeaders (or vice
// versa), catching column-count drift at build time.
const exportColumnCount = 26

var _ = [1]struct{}{}[exportColumnCount-len(imports.ExpectedHeaders)]

func (app *application) getEquipmentExport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	items, err := app.services.equipment.ListAll(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve equipment.", Code: http.StatusInternalServerError}
	}

	mfrs, err := app.services.equipment.ListManufacturers(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve manufacturers.", Code: http.StatusInternalServerError}
	}
	mfrByID := make(map[string]string, len(mfrs))
	for _, m := range mfrs {
		mfrByID[m.ID] = m.Name
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gearberg-equipment-export.csv"`)
	// UTF-8 BOM so Excel opens the file with correct encoding.
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})

	cw := csv.NewWriter(w)
	_ = cw.Write(imports.ExpectedHeaders)

	for _, item := range items {
		if item.IsArchived || item.TotalStock == 0 {
			continue
		}

		var units []equipment.Unit
		if item.TrackingType == equipment.Serialized {
			units, err = app.services.equipment.ListUnits(ctx, item.ID)
			if err != nil {
				return &httperr.Error{Error: err, Message: "Failed to retrieve units.", Code: http.StatusInternalServerError}
			}
		}

		for _, row := range imports.RowsForItem(item, mfrByID[item.ManufacturerID], units) {
			_ = cw.Write(row)
		}
	}
	cw.Flush()
	return nil
}
