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
// status and error message and continues. Steps that need DB access close
// over their own dependencies.
type Step func(ctx context.Context, rows []Row) ([]Row, error)

// CommitHandler writes committed rows to the target domain (e.g. equipment).
type CommitHandler interface {
	Commit(ctx context.Context, orgID string, rows []Row, mappings []Mapping) error
}

// StepProvider returns the validation steps for a given org.
type StepProvider interface {
	StepsFor(orgID string) []Step
}

// Service orchestrates the import pipeline.
type Service struct {
	db      *sql.DB
	repo    *Repository
	steps   StepProvider
	handler CommitHandler
	entity  string
}

// NewService returns a new Service.
func NewService(db *sql.DB, repo *Repository, steps StepProvider, handler CommitHandler, entity string) *Service {
	return &Service{db: db, repo: repo, steps: steps, handler: handler, entity: entity}
}

// StartSession parses the file, creates an import session, stores raw rows,
// runs validation steps, and returns the staged Session.
func (s *Service) StartSession(ctx context.Context, orgID string, format Format, r io.Reader) (Session, error) {
	reader, err := readerFor(format)
	if err != nil {
		return Session{}, fmt.Errorf("StartSession: %w", err)
	}
	records, err := reader.Read(ctx, r)
	if err != nil {
		return Session{}, fmt.Errorf("StartSession: read: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("StartSession: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	session, rows, err := s.stageSession(ctx, tx, orgID, format, records)
	if err != nil {
		return Session{}, fmt.Errorf("StartSession: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("StartSession: commit: %w", err)
	}

	rows, err = validateRows(ctx, s.steps.StepsFor(orgID), rows)
	if err != nil {
		return Session{}, fmt.Errorf("StartSession: %w", err)
	}

	if err := persistRows(ctx, s.repo, rows); err != nil {
		return Session{}, fmt.Errorf("StartSession: %w", err)
	}

	session, err = s.repo.UpdateSessionStatus(ctx, session.ID, StatusReady)
	if err != nil {
		return Session{}, fmt.Errorf("StartSession: %w", err)
	}

	return session, nil
}

func (s *Service) stageSession(ctx context.Context, tx *sql.Tx, orgID string, format Format, records []RawRecord) (Session, []Row, error) {
	session, err := s.repo.InsertSessionTx(ctx, tx, Session{
		ID:           uid.New(),
		OrgID:        orgID,
		Format:       format,
		Status:       StatusDraft,
		TargetEntity: s.entity,
	})
	if err != nil {
		return Session{}, nil, fmt.Errorf("stageSession: %w", err)
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
				return Session{}, nil, fmt.Errorf("stageSession: auto-map %q: %w", col, err)
			}
		}
	}

	rows := make([]Row, 0, len(records))
	for _, rec := range records {
		blob, err := json.Marshal(rec.Fields)
		if err != nil {
			return Session{}, nil, fmt.Errorf("stageSession: marshal row %d: %w", rec.RowNumber, err)
		}
		row, err := s.repo.InsertDataTx(ctx, tx, Row{
			ID:        uid.New(),
			SessionID: session.ID,
			RowNumber: int64(rec.RowNumber),
			Data:      string(blob),
		})
		if err != nil {
			return Session{}, nil, fmt.Errorf("stageSession: insert row %d: %w", rec.RowNumber, err)
		}
		rows = append(rows, row)
	}

	return session, rows, nil
}

// validateRows runs all steps against rows and assigns ActionCreate to rows
// that passed all steps.
func validateRows(ctx context.Context, steps []Step, rows []Row) ([]Row, error) {
	rows, err := runSteps(ctx, rows, steps)
	if err != nil {
		return nil, fmt.Errorf("ValidateRows: %w", err)
	}
	for i, row := range rows {
		if row.Status == RowStatusNew {
			rows[i].Status = RowStatusValid
			rows[i].Action = ActionCreate
		}
	}
	return rows, nil
}

// persistRows writes the status and action of each row back to the database.
func persistRows(ctx context.Context, repo *Repository, rows []Row) error {
	for _, row := range rows {
		if err := repo.UpdateDataStatus(ctx, row.ID, row.Status, row.ErrorMessage); err != nil {
			return fmt.Errorf("PersistRows: row %s: %w", row.ID, err)
		}
		if _, err := repo.UpdateDataAction(ctx, row.ID, row.Action); err != nil {
			return fmt.Errorf("PersistRows: action %s: %w", row.ID, err)
		}
	}
	return nil
}

// GetStagedSession returns the active staged session for an org, or ErrNotFound if none exists.
func (s *Service) GetStagedSession(ctx context.Context, orgID string) (Session, error) {
	session, err := s.repo.GetStagedSession(ctx, orgID)
	if err != nil {
		return Session{}, fmt.Errorf("Get: %w", err)
	}
	return session, nil
}

// Delete removes a session and all its rows and mappings (via CASCADE).
func (s *Service) Delete(ctx context.Context, sessionID string) error {
	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
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

// Commit finalises the import by delegating target-entity writes to the
// CommitHandler provided at construction.
func (s *Service) Commit(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	rows, err := s.repo.ListDataByAction(ctx, sessionID, ActionCreate)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	mappings, err := s.repo.ListMappings(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	if err := s.handler.Commit(ctx, session.OrgID, rows, mappings); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	if _, err := s.repo.UpdateSessionStatus(ctx, sessionID, StatusCommitted); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	return nil
}

func runSteps(ctx context.Context, rows []Row, steps []Step) ([]Row, error) {
	var err error
	for _, step := range steps {
		rows, err = step(ctx, rows)
		if err != nil {
			return nil, err
		}
	}
	return rows, nil
}
