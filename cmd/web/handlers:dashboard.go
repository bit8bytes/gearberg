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
	"fmt"
	"net/http"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

type dashboardData struct {
	OrgID                string
	TotalValue           string
	TotalStock           int64
	EquipmentOverdue     int64
	EquipmentOverdueSoon int64
}

func (app *application) getDashboard(w http.ResponseWriter, r *http.Request) *httperr.Error {
	orgID := r.PathValue("org_id")

	orgSettings, err := app.services.orgsettings.Get(r.Context(), orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	stats, err := app.services.equipment.Stats(r.Context(), orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency.Symbol()
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = dashboardData{
		OrgID:                orgID,
		TotalValue:           fmt.Sprintf("%s %.2f", currency, float64(stats.TotalValue)/100),
		TotalStock:           stats.TotalStock,
		EquipmentOverdue:     stats.EquipmentOverdue,
		EquipmentOverdueSoon: stats.EquipmentOverdueSoon,
	}
	return app.html.Render(w, r, http.StatusOK, pages.Dashboard, tmplData)
}
