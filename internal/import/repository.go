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
	"errors"
	"fmt"

	gendata "github.com/bit8bytes/gearberg/internal/database/queries/gen/importdata"
	genmappings "github.com/bit8bytes/gearberg/internal/database/queries/gen/importmappings"
	gensessions "github.com/bit8bytes/gearberg/internal/database/queries/gen/importsessions"
)

// Repository provides data access for import sessions, rows, and mappings.
type Repository struct {
	db       *sql.DB
	sessions *gensessions.Queries
	data     *gendata.Queries
	mappings *genmappings.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:       db,
		sessions: gensessions.New(db),
		data:     gendata.New(db),
		mappings: genmappings.New(db),
	}
}

// InsertSessionTx inserts a new import session within an existing transaction.
func (r *Repository) InsertSessionTx(ctx context.Context, tx *sql.Tx, s Session) (Session, error) {
	row, err := r.sessions.WithTx(tx).InsertSession(ctx, gensessions.InsertSessionParams{
		ID:           s.ID,
		OrgID:        s.OrgID,
		Format:       string(s.Format),
		Status:       string(s.Status),
		TargetEntity: s.TargetEntity,
	})
	if err != nil {
		return Session{}, fmt.Errorf("InsertSessionTx: %w", err)
	}
	return sessionFromRecord(row), nil
}

// GetSession returns a session by ID, or ErrNotFound if none exists.
func (r *Repository) GetSession(ctx context.Context, id string) (Session, error) {
	row, err := r.sessions.GetSession(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("GetSession: %w", err)
	}
	return sessionFromRecord(row), nil
}

// GetStagedSession returns the active ready session for an org, or ErrNotFound if none exists.
func (r *Repository) GetStagedSession(ctx context.Context, orgID string) (Session, error) {
	row, err := r.sessions.GetStagedSession(ctx, gensessions.GetStagedSessionParams{
		OrgID:  orgID,
		Status: string(StatusReady),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("GetStagedSession: %w", err)
	}
	return sessionFromRecord(row), nil
}

// DeleteSession deletes a session and cascades to its rows and mappings.
func (r *Repository) DeleteSession(ctx context.Context, id string) error {
	if err := r.sessions.DeleteSession(ctx, id); err != nil {
		return fmt.Errorf("DeleteSession: %w", err)
	}
	return nil
}

// UpdateSessionStatus transitions a session to the given status and returns the updated record.
func (r *Repository) UpdateSessionStatus(ctx context.Context, id string, status Status) (Session, error) {
	row, err := r.sessions.UpdateSessionStatus(ctx, gensessions.UpdateSessionStatusParams{
		Status: string(status),
		ID:     id,
	})
	if err != nil {
		return Session{}, fmt.Errorf("UpdateSessionStatus: %w", err)
	}
	return sessionFromRecord(row), nil
}

// InsertDataTx inserts a staged import row within an existing transaction.
func (r *Repository) InsertDataTx(ctx context.Context, tx *sql.Tx, row Row) (Row, error) {
	rec, err := r.data.WithTx(tx).InsertData(ctx, gendata.InsertDataParams{
		ID:        row.ID,
		SessionID: row.SessionID,
		RowNumber: row.RowNumber,
		Data:      row.Data,
	})
	if err != nil {
		return Row{}, fmt.Errorf("InsertDataTx: %w", err)
	}
	return dataFromRecord(rec), nil
}

// ListData returns all staged rows for a session.
func (r *Repository) ListData(ctx context.Context, sessionID string) ([]Row, error) {
	recs, err := r.data.ListData(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("ListData: %w", err)
	}
	rows := make([]Row, len(recs))
	for i, rec := range recs {
		rows[i] = dataFromRecord(rec)
	}
	return rows, nil
}

// ListDataByAction returns staged rows filtered by the given action.
func (r *Repository) ListDataByAction(ctx context.Context, sessionID string, action Action) ([]Row, error) {
	recs, err := r.data.ListDataByAction(ctx, gendata.ListDataByActionParams{
		SessionID: sessionID,
		Action:    string(action),
	})
	if err != nil {
		return nil, fmt.Errorf("ListDataByAction: %w", err)
	}
	rows := make([]Row, len(recs))
	for i, rec := range recs {
		rows[i] = dataFromRecord(rec)
	}
	return rows, nil
}

// UpdateDataStatus sets the validation status and error message for a row.
func (r *Repository) UpdateDataStatus(ctx context.Context, id string, status RowStatus, errMsg string) error {
	_, err := r.data.UpdateDataStatus(ctx, gendata.UpdateDataStatusParams{
		Status:       string(status),
		ErrorMessage: errMsg,
		ID:           id,
	})
	if err != nil {
		return fmt.Errorf("UpdateDataStatus: %w", err)
	}
	return nil
}

// UpdateDataAction sets the user-chosen action for a row and returns the updated record.
func (r *Repository) UpdateDataAction(ctx context.Context, id string, action Action) (Row, error) {
	rec, err := r.data.UpdateDataAction(ctx, gendata.UpdateDataActionParams{
		Action: string(action),
		ID:     id,
	})
	if err != nil {
		return Row{}, fmt.Errorf("UpdateDataAction: %w", err)
	}
	return dataFromRecord(rec), nil
}

// InsertMappingTx inserts a column mapping within an existing transaction.
func (r *Repository) InsertMappingTx(ctx context.Context, tx *sql.Tx, m Mapping) (Mapping, error) {
	rec, err := r.mappings.WithTx(tx).InsertMapping(ctx, genmappings.InsertMappingParams{
		ID:          m.ID,
		SessionID:   m.SessionID,
		SourceCol:   m.SourceCol,
		TargetField: m.TargetField,
	})
	if err != nil {
		return Mapping{}, fmt.Errorf("InsertMappingTx: %w", err)
	}
	return mappingFromRecord(rec), nil
}

// ListMappings returns all column mappings for a session.
func (r *Repository) ListMappings(ctx context.Context, sessionID string) ([]Mapping, error) {
	recs, err := r.mappings.ListMappings(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("ListMappings: %w", err)
	}
	ms := make([]Mapping, len(recs))
	for i, rec := range recs {
		ms[i] = mappingFromRecord(rec)
	}
	return ms, nil
}

func sessionFromRecord(r gensessions.ImportSession) Session {
	return Session{
		ID:           r.ID,
		OrgID:        r.OrgID,
		Format:       Format(r.Format),
		Status:       Status(r.Status),
		TargetEntity: r.TargetEntity,
		CreatedAt:    r.CreatedAt,
	}
}

func dataFromRecord(r gendata.ImportDatum) Row {
	return Row{
		ID:           r.ID,
		SessionID:    r.SessionID,
		RowNumber:    r.RowNumber,
		Data:         r.Data,
		Status:       RowStatus(r.Status),
		ErrorMessage: r.ErrorMessage,
		Action:       Action(r.Action),
	}
}

func mappingFromRecord(r genmappings.ImportMapping) Mapping {
	return Mapping{
		ID:          r.ID,
		SessionID:   r.SessionID,
		SourceCol:   r.SourceCol,
		TargetField: r.TargetField,
	}
}
