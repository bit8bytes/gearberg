package main

import (
	"encoding/csv"
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
			item.Pricing.RentalPrice.ToDecimal(),
			item.Pricing.PurchasePrice.ToDecimal(),
			item.Notes,
			item.Properties.Weight.ToKG(),
			item.Properties.Width.ToCM(),
			item.Properties.Height.ToCM(),
			item.Properties.Depth.ToCM(),
			item.Properties.Voltage.ToV(),
			item.Properties.Current.ToA(),
			item.Properties.Power.ToW(),
			item.Properties.WireGauge.String(),
			strconv.FormatInt(item.TotalStock, 10),
		})
	}
	cw.Flush()
	return nil
}
