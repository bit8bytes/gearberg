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

package imports

import (
	"context"
	_ "embed"
	"fmt"
	"io"

	pkgcsv "github.com/bit8bytes/gearberg/pkg/csv"
)

//go:embed template.csv
var TemplateCSV []byte

// CSVReader adapts pkg/csv.Reader to the imports.Reader and imports.Inspector interfaces.
type CSVReader struct {
	r pkgcsv.Reader
}

func init() {
	RegisterFormat(FormatCSV, &CSVReader{r: pkgcsv.Reader{}})
}

// NewCSVReader returns a CSVReader wrapping a pkg/csv.Reader.
func NewCSVReader(r pkgcsv.Reader) *CSVReader {
	return &CSVReader{r: r}
}

func (c *CSVReader) Read(ctx context.Context, r io.Reader) ([]RawRecord, error) {
	recs, err := c.r.Read(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("CSVReader.Read: %w", err)
	}
	out := make([]RawRecord, len(recs))
	for i, rec := range recs {
		out[i] = RawRecord{
			RowNumber: rec.Line,
			Fields:    rec.Fields,
		}
	}
	return out, nil
}

func (c *CSVReader) InspectHeaders(ctx context.Context, r io.Reader) ([]string, error) {
	headers, err := c.r.InspectHeaders(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("CSVReader.InspectHeaders: %w", err)
	}
	return headers, nil
}
