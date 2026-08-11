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
	"strings"
	"testing"

	"github.com/bit8bytes/gearberg/internal/equipment/imports"
)

func TestEquipmentImportExportRoundTrip(t *testing.T) {
	alice.signin(t)
	importID := roundTripUpload(t)
	roundTripConfirm(t, importID)
	body := roundTripExport(t)
	alice.signout(t)
	assertRoundTrip(t, body)
}

func roundTripUpload(t *testing.T) string {
	t.Helper()
	code, header, _ := ts.postFile(t, "/orgs/"+alice.orgID+"/equipment/import", "file", "template.csv", imports.TemplateCSV)
	if code != http.StatusSeeOther {
		t.Fatalf("upload: want 303, got %d", code)
	}
	return importIDFromLocation(t, header.Get("Location"))
}

func roundTripConfirm(t *testing.T, importID string) {
	t.Helper()
	code, _, _ := ts.postForm(t, "/orgs/"+alice.orgID+"/equipment/import/confirm", url.Values{"import_id": {importID}})
	if code != http.StatusSeeOther {
		t.Fatalf("confirm: want 303, got %d", code)
	}
}

func roundTripExport(t *testing.T) []byte {
	t.Helper()
	code, _, body := ts.get(t, "/orgs/"+alice.orgID+"/equipment/export")
	if code != http.StatusOK {
		t.Fatalf("export: want 200, got %d", code)
	}
	return body
}

func assertRoundTrip(t *testing.T, body []byte) {
	t.Helper()
	rows := parseExportCSV(t, body)
	assertExportCounts(t, rows)
	assertShureQuantity(t, rows)
	assertSonySerials(t, rows)
}

func parseExportCSV(t *testing.T, body []byte) []imports.RawRow {
	t.Helper()
	// Strip the UTF-8 BOM the handler prepends so ParseCSV sees clean bytes.
	if len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF {
		body = body[3:]
	}
	rows, err := imports.ParseCSV(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("ParseCSV on export: %v", err)
	}
	return rows
}

func assertExportCounts(t *testing.T, rows []imports.RawRow) {
	t.Helper()
	counts := make(map[string]int)
	for _, r := range rows {
		counts[r.Name]++
	}
	if counts["Shure SM58"] != 1 {
		t.Errorf("Shure SM58: want 1 export row, got %d", counts["Shure SM58"])
	}
	if counts["Sony A7 IV"] != 2 {
		t.Errorf("Sony A7 IV: want 2 export rows (one per unit), got %d", counts["Sony A7 IV"])
	}
	if counts["Pelican 1510 Case"] != 1 {
		t.Errorf("Pelican 1510 Case: want 1 export row, got %d", counts["Pelican 1510 Case"])
	}
}

func assertShureQuantity(t *testing.T, rows []imports.RawRow) {
	t.Helper()
	for _, r := range rows {
		if r.Name == "Shure SM58" && r.Quantity != "7" {
			t.Errorf("Shure SM58: want quantity 7, got %q", r.Quantity)
		}
	}
}

func assertSonySerials(t *testing.T, rows []imports.RawRow) {
	t.Helper()
	serials := make(map[string]bool)
	for _, r := range rows {
		if r.Name == "Sony A7 IV" {
			serials[r.UnitSerialNumber] = true
		}
	}
	for _, want := range []string{"SN-A7IV-001", "SN-A7IV-002"} {
		if !serials[want] {
			t.Errorf("Sony A7 IV: serial %q not found in export", want)
		}
	}
}

// importIDFromLocation extracts the ?id= query parameter from a redirect Location header.
func importIDFromLocation(t *testing.T, loc string) string {
	t.Helper()
	for part := range strings.SplitSeq(loc, "?") {
		params, _ := url.ParseQuery(part)
		if id := params.Get("id"); id != "" {
			return id
		}
	}
	t.Fatalf("redirect location %q contains no import id", loc)
	return ""
}
