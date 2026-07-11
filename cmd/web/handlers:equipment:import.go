package main

import (
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/equipment/imports"
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
	Rows       []imports.Row
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

	reRender := func(msg string) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Data = equipmentImportData{OrgID: orgID, Error: msg}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentImport, data)
	}

	const maxImportBytes = 32 << 20                              // 32 MiB
	if err := r.ParseMultipartForm(maxImportBytes); err != nil { //nolint:gosec // maxImportBytes is a bounded constant (32 MiB)
		return reRender("Could not parse form.")
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		return reRender("No file uploaded.")
	}
	defer func() { _ = f.Close() }()

	rawRows, parseErr := imports.ParseCSV(f)
	if parseErr != nil {
		return reRender(parseErr.Error())
	}

	importID, err := app.services.equipmentImports.Stage(ctx, orgID, rawRows)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to stage import.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/import?id="+url.QueryEscape(importID), http.StatusSeeOther)
	return nil
}

func (app *application) renderImportPreview(w http.ResponseWriter, r *http.Request, orgID, importID string) *httperr.Error {
	ctx := r.Context()

	rows, err := app.services.equipmentImports.ListStaged(ctx, importID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to load import preview.", Code: http.StatusInternalServerError}
	}

	var cntNew, cntError int
	for _, row := range rows {
		switch row.Status {
		case imports.StatusNew:
			cntNew++
		case imports.StatusError:
			cntError++
		}
	}

	data := app.html.TemplateData(r)
	data.Data = equipmentImportPreviewData{
		OrgID:      orgID,
		ImportID:   importID,
		Rows:       rows,
		CountNew:   cntNew,
		CountError: cntError,
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentImportPreview, data)
}

func (app *application) postEquipmentImportConfirm(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	if err := r.ParseForm(); err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}
	importID := r.FormValue("import_id")

	if err := app.services.equipmentImports.Commit(ctx, importID, orgID); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to commit import.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment", http.StatusSeeOther)
	return nil
}

// getEquipmentImportTemplate serves a ready-to-fill CSV template for download.
func (app *application) getEquipmentImportTemplate(w http.ResponseWriter, _ *http.Request) *httperr.Error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gearberg-import-template.csv"`)
	_, _ = w.Write(imports.TemplateCSV)
	return nil
}
