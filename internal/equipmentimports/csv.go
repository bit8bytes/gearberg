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

package equipmentimports

import (
	"context"
	"fmt"
	"io"
	"strings"

	pkgcsv "github.com/bit8bytes/gearberg/pkg/csv"
)

// columnAliases maps legacy column names to their current canonical name.
// Old exports that used a different name for a column are accepted transparently.
var columnAliases = map[string]string{
	// "Has Content" was a boolean column (TRUE/FALSE) replaced by "Equipment Type"
	// (Standard/Kit). Values are normalised in MapRecords.
	"Has Content": "Equipment Type",
}

// ParseCSV reads a CSV (with or without a UTF-8 BOM) and returns processed rows
// with all values already converted to DB units (cents, grams, millimetres, etc.).
// All columns in ExpectedHeaders must be present; order and extra columns are ignored.
// Every row is initialised to StateValid; ImportID and OrgID are left empty —
// Stage sets them before persisting.
func ParseCSV(r io.Reader) ([]ProcessedRow, error) {
	rd := &pkgcsv.Reader{Aliases: columnAliases}
	records, err := rd.Read(context.Background(), r)
	if err != nil {
		return nil, fmt.Errorf("ParseCSV: %w", err)
	}
	for _, name := range ExpectedHeaders {
		if _, ok := records[0].Fields[name]; !ok {
			return nil, fmt.Errorf("ParseCSV: missing required column %q", name)
		}
	}
	return MapRecords(records, "", ""), nil
}

// normalizeEquipmentTypeLabel maps legacy boolean values from the old "Has Content"
// column to the current Equipment Type labels used by TypeFromString.
func normalizeEquipmentTypeLabel(v string) string {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "TRUE", "1":
		return "Kit"
	case "FALSE", "0":
		return "Standard"
	default:
		return v
	}
}
