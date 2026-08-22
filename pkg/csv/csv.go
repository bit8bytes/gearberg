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

// Package csv parses CSV files into generic records.
package csv

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// Record is one parsed row: a line number and its field values keyed by column header.
type Record struct {
	Line   int
	Fields map[string]string
}

// Reader parses a CSV file into records.
// Column order is irrelevant; fields are keyed by their header name.
// Extra columns are ignored; a UTF-8 BOM is stripped automatically.
type Reader struct {
	// Aliases maps legacy column names to their canonical header name.
	// Useful for accepting old exports that renamed a column.
	Aliases map[string]string
}

// InspectHeaders reads only the first line of r and returns the column names.
// Used by the field-mapping UI step before a full parse is performed.
func (rd *Reader) InspectHeaders(_ context.Context, r io.Reader) ([]string, error) {
	cr, err := rd.newCSVReader(r)
	if err != nil {
		return nil, err
	}
	rawHeader, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("csv: read header: %w", err)
	}
	headers := make([]string, len(rawHeader))
	for i, h := range rawHeader {
		canonical := h
		if rd.Aliases != nil {
			if alias, ok := rd.Aliases[h]; ok {
				canonical = alias
			}
		}
		headers[i] = canonical
	}
	return headers, nil
}

// Read parses r and returns one Record per data row.
// Returns an error if the file has no data rows.
func (rd *Reader) Read(_ context.Context, r io.Reader) ([]Record, error) {
	cr, err := rd.newCSVReader(r)
	if err != nil {
		return nil, err
	}

	idx, err := rd.buildIndex(cr)
	if err != nil {
		return nil, err
	}

	return readRecords(cr, idx)
}

func (rd *Reader) newCSVReader(r io.Reader) (*csv.Reader, error) {
	br := bufio.NewReader(r)
	if peek, err := br.Peek(3); err == nil && peek[0] == 0xEF && peek[1] == 0xBB && peek[2] == 0xBF {
		_, _ = br.Discard(3)
	}
	cr := csv.NewReader(br)
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1
	return cr, nil
}

func (rd *Reader) buildIndex(cr *csv.Reader) (map[string]int, error) {
	rawHeader, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("csv: read header: %w", err)
	}
	idx := make(map[string]int, len(rawHeader))
	for i, h := range rawHeader {
		canonical := h
		if rd.Aliases != nil {
			if alias, ok := rd.Aliases[h]; ok {
				canonical = alias
			}
		}
		idx[canonical] = i
	}
	return idx, nil
}

func readRecords(cr *csv.Reader, idx map[string]int) ([]Record, error) {
	var records []Record
	line := 1
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv: line %d: %w", line, err)
		}
		line++
		fields := make(map[string]string, len(idx))
		for name, i := range idx {
			if i < len(row) {
				fields[name] = strings.TrimSpace(row[i])
			}
		}
		records = append(records, Record{Line: line, Fields: fields})
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("csv: no data rows")
	}
	return records, nil
}
