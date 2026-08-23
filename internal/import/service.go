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
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bit8bytes/gearberg/internal/uid"
)

// Reader parses an input source into raw records.
// Satisfied by *csv.Reader or any future format reader.
type Reader interface {
	Read(ctx context.Context, r io.Reader) ([]RawRecord, error)
}

// Inspector reads only the header layer of a file to discover column names
// for the field-mapping UI, without parsing the full dataset.
type Inspector interface {
	InspectHeaders(ctx context.Context, r io.Reader) ([]string, error)
}

// Step is a single validation stage applied to a batch of rows.
// A Step must never return early on a per-row error; it sets the row's
// status and error message and continues.
type Step func(ctx context.Context, repo *Repository, rows []Row) ([]Row, error)

// Service orchestrates the import pipeline.
type Service struct {
	db    *sql.DB
	repo  *Repository
	steps []Step
}

// NewService returns a new Service with the given validation steps.
func NewService(db *sql.DB, repo *Repository, steps ...Step) *Service {
	return &Service{db: db, repo: repo, steps: steps}
}

// NewSession parses the file, creates an import session, stores raw rows,
// and runs validation. Returns the staged Session.
func (s *Service) NewSession(ctx context.Context, orgID string, format Format, targetEntity string, r io.Reader, reader Reader) (Session, error) {
	records, err := reader.Read(ctx, r)
	if err != nil {
		return Session{}, fmt.Errorf("NewSession: read: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("NewSession: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	session, err := s.repo.InsertSessionTx(ctx, tx, Session{
		ID:           uid.New(),
		OrgID:        orgID,
		Format:       format,
		Status:       StatusPending,
		TargetEntity: targetEntity,
	})
	if err != nil {
		return Session{}, fmt.Errorf("NewSession: %w", err)
	}

	// Auto-map source col → target field using the first record's keys.
	// The mapping UI is not implemented yet, so we assume the source columns
	// already match the internal field names (e.g. via the Gearberg template CSV).
	// When the mapping UI is added, this block is replaced by user-defined mappings.
	if len(records) > 0 {
		for col := range records[0].Fields {
			if _, err := s.repo.InsertMappingTx(ctx, tx, Mapping{
				ID:          uid.New(),
				SessionID:   session.ID,
				SourceCol:   col,
				TargetField: col,
			}); err != nil {
				return Session{}, fmt.Errorf("NewSession: auto-map %q: %w", col, err)
			}
		}
	}

	rows := make([]Row, 0, len(records))
	for _, rec := range records {
		blob, err := json.Marshal(rec.Fields)
		if err != nil {
			return Session{}, fmt.Errorf("NewSession: marshal row %d: %w", rec.RowNumber, err)
		}
		row, err := s.repo.InsertDataTx(ctx, tx, Row{
			ID:        uid.New(),
			SessionID: session.ID,
			RowNumber: int64(rec.RowNumber),
			Data:      string(blob),
		})
		if err != nil {
			return Session{}, fmt.Errorf("NewSession: insert row %d: %w", rec.RowNumber, err)
		}
		rows = append(rows, row)
	}

	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("NewSession: commit: %w", err)
	}

	rows, err = s.runSteps(ctx, rows)
	if err != nil {
		return Session{}, fmt.Errorf("NewSession: validate: %w", err)
	}

	session, err = s.repo.UpdateSessionStatus(ctx, session.ID, StatusStaged)
	if err != nil {
		return Session{}, fmt.Errorf("NewSession: %w", err)
	}

	return session, nil
}

// ListData returns all staged rows for a session.
func (s *Service) ListData(ctx context.Context, sessionID string) ([]Row, error) {
	rows, err := s.repo.ListData(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("ListData: %w", err)
	}
	return rows, nil
}

// Review applies user decisions to staged rows.
func (s *Service) Review(ctx context.Context, decisions []Decision) error {
	for _, d := range decisions {
		if _, err := s.repo.UpdateDataAction(ctx, d.RowID, d.Action); err != nil {
			return fmt.Errorf("Review: row %s: %w", d.RowID, err)
		}
	}
	return nil
}

// Commit finalises the import. Target-entity writes are delegated to the
// caller via a CommitHandler so this package stays format- and domain-agnostic.
func (s *Service) Commit(ctx context.Context, sessionID string, handler CommitHandler) error {
	rows, err := s.repo.ListDataByAction(ctx, sessionID, ActionCreate)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	mappings, err := s.repo.ListMappings(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	if err := handler.Commit(ctx, rows, mappings); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	if _, err := s.repo.UpdateSessionStatus(ctx, sessionID, StatusCommitted); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	return nil
}

// CommitHandler writes committed rows to the target domain (e.g. equipment).
type CommitHandler interface {
	Commit(ctx context.Context, rows []Row, mappings []Mapping) error
}

func (s *Service) runSteps(ctx context.Context, rows []Row) ([]Row, error) {
	var err error
	for _, step := range s.steps {
		rows, err = step(ctx, s.repo, rows)
		if err != nil {
			return nil, err
		}
	}
	return rows, nil
}
