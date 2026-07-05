package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
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

	rawRows, parseErr := parseImportCSV(f)
	if parseErr != "" {
		return reRender("Failed to parse CSV.")
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

const importTemplateCSV = "Name,Type,Usage,Category,Manufacturer,Location,Rental Price,Resale Price,Notes,Weight (g),Width (mm),Height (mm),Depth (mm),Voltage (V),Current (mA),Power (mW),Wire Gauge (mm² ×100),Quantity\n" +
	"Shure SM58,Bulk,Rental,Audio,Shure,Main Warehouse,15.00,99.00,Cardioid dynamic vocal microphone,298,47,47,162,,,,,7\n" +
	"Sony SRS-XB43,Serialized,Rental,Audio,Sony,Main Warehouse,25.00,180.00,Portable Bluetooth speaker with extra bass,900,220,220,95,5,2400,12000,,4\n"

// getEquipmentImportTemplate serves a ready-to-fill CSV template for download.
func (app *application) getEquipmentImportTemplate(w http.ResponseWriter, _ *http.Request) *httperr.Error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gearberg-import-template.csv"`)
	_, _ = w.Write([]byte(importTemplateCSV))
	return nil
}

// parseImportCSV reads a CSV file and returns RawRows. Returns an error message on failure.
func parseImportCSV(r io.Reader) ([]imports.RawRow, string) {
	br := bufio.NewReader(r)
	// Strip UTF-8 BOM produced by the export so round-tripped files parse cleanly.
	if peek, err := br.Peek(3); err == nil && peek[0] == 0xEF && peek[1] == 0xBB && peek[2] == 0xBF {
		_, _ = br.Discard(3)
	}
	cr := csv.NewReader(br)
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		return nil, "Could not read CSV file."
	}
	if msg := validateImportHeader(header); msg != "" {
		return nil, msg
	}
	return readImportRows(cr)
}

func validateImportHeader(header []string) string {
	if len(header) != len(imports.ExpectedHeaders) {
		return fmt.Sprintf("Expected %d columns, got %d. Make sure the file matches the export format.", len(imports.ExpectedHeaders), len(header))
	}
	for i, h := range header {
		if h != imports.ExpectedHeaders[i] {
			return fmt.Sprintf("Column %d: expected %q, got %q.", i+1, imports.ExpectedHeaders[i], h)
		}
	}
	return ""
}

// readImportRows reads data rows from a CSV reader after the header has been consumed.
// Column order matches imports.ExpectedHeaders exactly.
func readImportRows(cr *csv.Reader) ([]imports.RawRow, string) {
	var rows []imports.RawRow
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "Error reading CSV file."
		}
		if len(record) < len(imports.ExpectedHeaders) {
			continue
		}
		rows = append(rows, imports.RawRow{
			Name:             strings.TrimSpace(record[0]),
			TypeLabel:        strings.TrimSpace(record[1]),
			UsageTypeLabel:   strings.TrimSpace(record[2]),
			CategoryName:     strings.TrimSpace(record[3]),
			ManufacturerName: strings.TrimSpace(record[4]),
			LocationName:     strings.TrimSpace(record[5]),
			RentalPrice:      strings.TrimSpace(record[6]),
			ResalePrice:      strings.TrimSpace(record[7]),
			Notes:            strings.TrimSpace(record[8]),
			WeightG:          strings.TrimSpace(record[9]),
			WidthMm:          strings.TrimSpace(record[10]),
			HeightMm:         strings.TrimSpace(record[11]),
			DepthMm:          strings.TrimSpace(record[12]),
			VoltageV:         strings.TrimSpace(record[13]),
			CurrentMa:        strings.TrimSpace(record[14]),
			PowerMw:          strings.TrimSpace(record[15]),
			WireGaugeMM2X100: strings.TrimSpace(record[16]),
			Quantity:         strings.TrimSpace(record[17]),
		})
	}
	if len(rows) == 0 {
		return nil, "The CSV file has no data rows."
	}
	return rows, ""
}
