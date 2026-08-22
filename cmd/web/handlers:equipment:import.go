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
	"net/url"

	"github.com/bit8bytes/gearberg/internal/equipmentimports"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

type equipmentImportData struct {
	OrgID string
	Error string
}

type equipmentImportPreviewData struct {
	OrgID      string
	ImportID   string
	Rows       []equipmentimports.GroupedRow
	CountNew   int
	CountError int
}

// getEquipmentImport serves the upload form when no ?id= param is present,
// or the staging preview when ?id= is set (after a successful upload).
func (app *application) getEquipmentImport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	orgID := r.PathValue("org_id")
	importID := r.URL.Query().Get("id")
	if importID == "" {
		data := app.html.TemplateData(r)
		data.Data = equipmentImportData{OrgID: orgID}
		return app.html.Render(w, r, http.StatusOK, pages.EquipmentImport, data)
	}
	return app.renderImportPreview(w, r, orgID, importID)
}

func (app *application) postEquipmentImport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	fail := func(msg string) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Data = equipmentImportData{OrgID: orgID, Error: msg}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentImport, data)
	}

	const maxImportBytes = 32 << 20                              // 32 MiB
	if err := r.ParseMultipartForm(maxImportBytes); err != nil { //nolint:gosec // maxImportBytes is a bounded constant (32 MiB)
		return fail("Could not parse form.")
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		return fail("No file uploaded.")
	}
	defer func() { _ = f.Close() }()

	rawRows, parseErr := equipmentimports.ParseCSV(f)
	if parseErr != nil {
		return fail(parseErr.Error())
	}

	importID, err := app.services.equipmentImports.Stage(ctx, orgID, rawRows)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/import?id="+url.QueryEscape(importID), http.StatusSeeOther)
	return nil
}

func (app *application) renderImportPreview(w http.ResponseWriter, r *http.Request, orgID, importID string) *httperr.Error {
	ctx := r.Context()

	staged, err := app.services.equipmentImports.ListStaged(ctx, importID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	previewRows, cntNew, cntError := equipmentimports.GroupRows(staged)

	data := app.html.TemplateData(r)
	data.Data = equipmentImportPreviewData{
		OrgID:      orgID,
		ImportID:   importID,
		Rows:       previewRows,
		CountNew:   cntNew,
		CountError: cntError,
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentImportPreview, data)
}

func (app *application) postEquipmentImportConfirm(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	if err := r.ParseForm(); err != nil {
		return httperr.BadRequest(err)
	}
	importID := r.FormValue("import_id")

	if err := app.services.equipmentImports.Commit(ctx, importID, orgID); err != nil {
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment", http.StatusSeeOther)
	return nil
}

// getEquipmentImportTemplate serves a ready-to-fill CSV template for download.
func (app *application) getEquipmentImportTemplate(w http.ResponseWriter, _ *http.Request) *httperr.Error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gearberg-import-template.csv"`)
	_, _ = w.Write(equipmentimports.TemplateCSV)
	return nil
}
