package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bit8bytes/gearberg/internal/equipment/imports"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

type equipmentImportData struct {
	OrgID string
	Error string
}

// importPreviewRow is a display-oriented view of a staged import row.
// Serialized rows with the same name are collapsed into a single entry;
// Stock holds the unit count for serialized items and quantity for bulk.
type importPreviewRow struct {
	RowNumber    int64
	Name         string
	TypeLabel    string
	CategoryName string
	Stock        string
	Status       string
	ErrorMessage string
}

type equipmentImportPreviewData struct {
	OrgID      string
	ImportID   string
	Rows       []importPreviewRow
	CountNew   int
	CountError int
}

// groupImportRows collapses serialized staging rows that share a name into one
// preview row and returns per-item counts for new and error items.
func groupImportRows(staged []imports.Row) (rows []importPreviewRow, cntNew, cntError int) {
	type group struct {
		row    importPreviewRow
		total  int
		hasErr bool
	}
	seen := make(map[string]*group)
	var order []string

	for _, r := range staged {
		if !strings.EqualFold(r.TypeLabel, "serialized") {
			pr := importPreviewRow{
				RowNumber:    r.RowNumber,
				Name:         r.Name,
				TypeLabel:    r.TypeLabel,
				CategoryName: r.CategoryName,
				Stock:        r.Quantity,
				Status:       r.Status,
				ErrorMessage: r.ErrorMessage,
			}
			rows = append(rows, pr)
			if r.Status == imports.StatusNew {
				cntNew++
			} else {
				cntError++
			}
			continue
		}

		key := strings.ToLower(r.Name)
		if _, ok := seen[key]; !ok {
			seen[key] = &group{row: importPreviewRow{
				RowNumber:    r.RowNumber,
				Name:         r.Name,
				TypeLabel:    r.TypeLabel,
				CategoryName: r.CategoryName,
				Status:       imports.StatusNew,
			}}
			order = append(order, key)
		}
		g := seen[key]
		g.total++
		if r.Status == imports.StatusError && !g.hasErr {
			g.hasErr = true
			g.row.Status = imports.StatusError
			g.row.ErrorMessage = r.ErrorMessage
		}
	}

	for _, key := range order {
		g := seen[key]
		g.row.Stock = fmt.Sprintf("%d", g.total)
		rows = append(rows, g.row)
		if g.hasErr {
			cntError++
		} else {
			cntNew++
		}
	}
	return
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

	staged, err := app.services.equipmentImports.ListStaged(ctx, importID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to load import preview.", Code: http.StatusInternalServerError}
	}

	previewRows, cntNew, cntError := groupImportRows(staged)

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
