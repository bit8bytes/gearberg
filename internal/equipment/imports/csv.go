package imports

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// ParseCSV reads a CSV (with or without a UTF-8 BOM) and returns the data rows.
// The header row must match ExpectedHeaders exactly.
func ParseCSV(r io.Reader) ([]RawRow, error) {
	br := bufio.NewReader(r)
	// Strip UTF-8 BOM produced by the export so round-tripped files parse cleanly.
	if peek, err := br.Peek(3); err == nil && peek[0] == 0xEF && peek[1] == 0xBB && peek[2] == 0xBF {
		_, _ = br.Discard(3)
	}
	cr := csv.NewReader(br)
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1 // allow variable field counts; short rows are padded in readRows

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("ParseCSV: read header: %w", err)
	}
	if err := validateHeader(header); err != nil {
		return nil, err
	}
	return readRows(cr)
}

func validateHeader(header []string) error {
	if len(header) != len(ExpectedHeaders) {
		return fmt.Errorf("expected %d columns, got %d", len(ExpectedHeaders), len(header))
	}
	for i, h := range header {
		if h != ExpectedHeaders[i] {
			return fmt.Errorf("column %d: expected %q, got %q", i+1, ExpectedHeaders[i], h)
		}
	}
	return nil
}

// readRows reads data rows after the header has been consumed.
// Column positions must match ExpectedHeaders exactly.
func readRows(cr *csv.Reader) ([]RawRow, error) {
	var rows []RawRow
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("readRows: %w", err)
		}
		if len(record) < len(ExpectedHeaders) {
			padded := make([]string, len(ExpectedHeaders))
			copy(padded, record)
			record = padded
		}
		rows = append(rows, RawRow{
			Name:                   strings.TrimSpace(record[0]),
			TypeLabel:              strings.TrimSpace(record[1]),
			UsageTypeLabel:         strings.TrimSpace(record[2]),
			CategoryName:           strings.TrimSpace(record[3]),
			ManufacturerName:       strings.TrimSpace(record[4]),
			LocationName:           strings.TrimSpace(record[5]),
			RentalPrice:            strings.TrimSpace(record[6]),
			ResalePrice:            strings.TrimSpace(record[7]),
			Notes:                  strings.TrimSpace(record[8]),
			WeightG:                strings.TrimSpace(record[9]),
			WidthMm:                strings.TrimSpace(record[10]),
			HeightMm:               strings.TrimSpace(record[11]),
			DepthMm:                strings.TrimSpace(record[12]),
			VoltageV:               strings.TrimSpace(record[13]),
			CurrentA:               strings.TrimSpace(record[14]),
			PowerW:                 strings.TrimSpace(record[15]),
			WireGaugeMM2X100:       strings.TrimSpace(record[16]),
			Quantity:               strings.TrimSpace(record[17]),
			HasContent:             strings.TrimSpace(record[18]),
			UnitSerialNumber:       strings.TrimSpace(record[19]),
			UnitManufacturerSerial: strings.TrimSpace(record[20]),
			UnitPurchasePrice:      strings.TrimSpace(record[21]),
			UnitPurchasedAt:        strings.TrimSpace(record[22]),
			NextInspectionAt:       strings.TrimSpace(record[23]),
			UnitIsActive:           strings.TrimSpace(record[24]),
			UnitRemark:             strings.TrimSpace(record[25]),
		})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("readRows: no data rows")
	}
	return rows, nil
}
