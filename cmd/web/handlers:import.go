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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/httperr"
	imports "github.com/bit8bytes/gearberg/internal/import"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	pkgcsv "github.com/bit8bytes/gearberg/pkg/csv"
)

const importMaxBytes = 32 << 20 // 32 MiB

type importUploadData struct {
	OrgID string
	Error string
}

type importReviewRow struct {
	ID           string
	RowNumber    int64
	Name         string
	Status       string
	ErrorMessage string
	Action       string
}

type importReviewData struct {
	OrgID        string
	SessionID    string
	Rows         []importReviewRow
	CountReady   int
	CountPending int
	CountError   int
}

// getImportTemplate serves the template CSV for download.
func (app *application) getImportTemplate(w http.ResponseWriter, r *http.Request) *httperr.Error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gearberg-import-template.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(imports.TemplateCSV)
	return nil
}

// getImport serves the upload form.
func (app *application) getImport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	orgID := r.PathValue("org_id")
	tmpl := app.html.TemplateData(r)
	tmpl.Data = importUploadData{OrgID: orgID}
	return app.html.Render(w, r, http.StatusOK, pages.ImportUpload, tmpl)
}

// postImport parses the uploaded file, creates an import session, and redirects
// to the review page.
func (app *application) postImport(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	if err := r.ParseMultipartForm(importMaxBytes); err != nil {
		return httperr.BadRequest(fmt.Errorf("file too large or malformed: %w", err))
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return httperr.BadRequest(fmt.Errorf("missing file: %w", err))
	}
	defer file.Close()

	reader := imports.NewCSVReader(pkgcsv.Reader{})
	session, err := app.services.imports.NewSession(ctx, orgID, imports.FormatCSV, "equipment", file, reader)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/import/"+url.PathEscape(session.ID), http.StatusSeeOther)
	return nil
}

// getImportReview shows staged rows for user review.
func (app *application) getImportReview(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	sessionID := r.PathValue("id")

	rows, err := app.services.imports.ListData(ctx, sessionID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := importReviewData{
		OrgID:     orgID,
		SessionID: sessionID,
		Rows:      make([]importReviewRow, 0, len(rows)),
	}

	for _, row := range rows {
		var fields map[string]string
		_ = json.Unmarshal([]byte(row.Data), &fields)
		name := fields["Name"]

		switch row.Status {
		case imports.RowStatusError:
			data.CountError++
		case imports.RowStatusNeedsReview:
			data.CountPending++
		default:
			if row.Action == imports.ActionSkip {
				data.CountPending++
			} else {
				data.CountReady++
			}
		}

		data.Rows = append(data.Rows, importReviewRow{
			ID:           row.ID,
			RowNumber:    row.RowNumber,
			Name:         name,
			Status:       string(row.Status),
			ErrorMessage: row.ErrorMessage,
			Action:       string(row.Action),
		})
	}

	tmpl := app.html.TemplateData(r)
	tmpl.Data = data
	return app.html.Render(w, r, http.StatusOK, pages.ImportReview, tmpl)
}

// postImportReview applies a single row decision (create/skip) and redirects back.
func (app *application) postImportReview(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	sessionID := r.PathValue("id")

	rowID := r.FormValue("row_id")
	action := imports.Action(r.FormValue("action"))

	if err := app.services.imports.Review(ctx, []imports.Decision{{RowID: rowID, Action: action}}); err != nil {
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/import/"+url.PathEscape(sessionID), http.StatusSeeOther)
	return nil
}

// postImportConfirm commits the import and redirects to the equipment list.
func (app *application) postImportConfirm(w http.ResponseWriter, r *http.Request) *httperr.Error {
	// TODO: implement CommitHandler for equipment
	orgID := r.PathValue("org_id")
	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment", http.StatusSeeOther)
	return nil
}
