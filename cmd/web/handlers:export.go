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

	"github.com/bit8bytes/gearberg/internal/httperr"
)

func (app *application) getExport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	orgID := r.PathValue("org_id")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gearberg-export.csv"`)
	if err := app.services.equipment.ExportCSV(r.Context(), w, orgID); err != nil {
		return httperr.InternalServerError(err)
	}
	return nil
}
