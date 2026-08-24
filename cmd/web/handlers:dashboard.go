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
	"net/http"

	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

type dashboardData struct {
	OrgID              string
	OverdueInspections []equipment.InspectionSummary
	SoonInspections    []equipment.InspectionSummary
}

func (app *application) getDashboard(w http.ResponseWriter, r *http.Request) *httperr.Error {
	orgID := r.PathValue("org_id")

	overdue, err := app.services.equipment.ListOverdueInspections(r.Context(), orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	soon, err := app.services.equipment.ListSoonInspections(r.Context(), orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = dashboardData{
		OrgID:              orgID,
		OverdueInspections: overdue,
		SoonInspections:    soon,
	}
	return app.html.Render(w, r, http.StatusOK, pages.Dashboard, tmplData)
}
