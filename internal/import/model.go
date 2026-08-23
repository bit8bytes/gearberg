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

// Package imports provides format-agnostic import functionality.
package imports

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("import: not found")

// Format identifies the source file format.
type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
)

var readers = map[Format]Reader{}

// RegisterFormat registers a Reader for the given Format.
// Call from an init() function in the format's package.
func RegisterFormat(f Format, r Reader) {
	readers[f] = r
}

func readerFor(f Format) (Reader, error) {
	r, ok := readers[f]
	if !ok {
		return nil, fmt.Errorf("import: no reader registered for format %q", f)
	}
	return r, nil
}

// Status tracks the lifecycle of an import session.
// Each value describes the state the session is in, not the action that caused it.
type Status string

const (
	StatusDraft     Status = "draft"     // uploaded, awaiting column mapping
	StatusMapped    Status = "mapped"    // mappings saved, awaiting validation
	StatusReady     Status = "ready"     // validated, awaiting user commit
	StatusCommitted Status = "committed" // import written to inventory
)

// Action is the user's decision for a single row.
type Action string

const (
	ActionPending Action = "pending"
	ActionCreate  Action = "create"
	ActionSkip    Action = "skip"
)

// RowStatus is the system's assessment of a single row.
type RowStatus string

const (
	RowStatusNew   RowStatus = "new"
	RowStatusValid RowStatus = "valid"
	RowStatusError RowStatus = "error"
)

// Session represents an import session.
type Session struct {
	ID           string
	OrgID        string
	Format       Format
	Status       Status
	TargetEntity string
	CreatedAt    int64
}

// Row is a single staged import row.
type Row struct {
	ID           string
	SessionID    string
	RowNumber    int64
	Data         string // JSON blob of raw source columns
	Status       RowStatus
	ErrorMessage string
	Action       Action
}

// Mapping is a user-defined column mapping for a session.
type Mapping struct {
	ID          string
	SessionID   string
	SourceCol   string
	TargetField string
}

// Decision is a user's action choice for a single row during review.
type Decision struct {
	RowID  string
	Action Action
}

// RawRecord is a parsed row from any source format.
// Fields maps source column names to their raw string values.
type RawRecord struct {
	RowNumber int
	Fields    map[string]string
}
