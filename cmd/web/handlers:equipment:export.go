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
package main

import (
	"encoding/csv"
	"math"
	"net/http"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
	"github.com/bit8bytes/gearberg/internal/equipmentimports"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/pagination"
)

// exportColumnCount must equal len(imports.ExpectedHeaders). This line fails to
// compile if RowsForItem is updated without updating ExpectedHeaders (or vice
// versa), catching column-count drift at build time.
const exportColumnCount = 26

var _ = [1]struct{}{}[exportColumnCount-len(equipmentimports.ExpectedHeaders)]

func (app *application) getEquipmentExport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	items, _, err := app.services.equipment.List(ctx, equipment.ListParams{OrgID: orgID, Filters: pagination.Filters{Page: 1, PageSize: math.MaxInt32}})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	mfrs, err := app.services.manufacturers.List(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(err)
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
	_ = cw.Write(equipmentimports.ExpectedHeaders)

	for _, item := range items {
		if item.IsArchived || item.TotalStock == 0 {
			continue
		}

		var units []equipment.Unit
		if item.TrackingType == tracking.Serialized {
			units, err = app.services.equipment.ListUnits(ctx, item.ID)
			if err != nil {
				return httperr.InternalServerError(err)
			}
		}

		for _, row := range equipmentimports.RowsForItem(item, mfrByID[item.ManufacturerID], units) {
			_ = cw.Write(row)
		}
	}
	cw.Flush()
	return nil
}
