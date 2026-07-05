package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bit8bytes/gearberg/internal/equipment/imports"
	"github.com/bit8bytes/gearberg/internal/httperr"
)

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
		_ = cw.Write([]string{
			item.Name,
			item.Type.Label(),
			item.UsageType.Label(),
			item.CategoryName,
			mfrByID[item.ManufacturerID],
			item.LocationName,
			exportPrice(item.RentalPrice),
			exportPrice(item.PurchasePrice),
			item.Notes,
			exportOptInt(item.WeightG),
			exportOptInt(item.WidthMM),
			exportOptInt(item.HeightMM),
			exportOptInt(item.DepthMM),
			exportOptInt(item.VoltageV),
			exportOptInt(item.CurrentMA),
			exportOptInt(item.PowerMW),
			exportOptInt(item.WireGaugeMM2X100),
			strconv.FormatInt(item.TotalStock, 10),
		})
	}
	cw.Flush()
	return nil
}

func exportPrice(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", float64(*v)/100)
}

func exportOptInt(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}
