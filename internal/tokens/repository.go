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

// Package tokens manages short-lived token records used for email verification
// and password reset flows.
package tokens

import (
	"context"
	"database/sql"
	"fmt"

	gentokens "github.com/bit8bytes/gearberg/internal/database/queries/gen/tokens"
)

// Record holds the account and scope associated with a verified token.
type Record struct {
	AccountID    string
	TokenScopeID int64
}

// Repository provides data access for tokens.
type Repository struct {
	tokens *gentokens.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db gentokens.DBTX) *Repository {
	return &Repository{
		tokens: gentokens.New(db),
	}
}

// CreateParams holds the fields needed to insert a new token.
type CreateParams struct {
	AccountID string
	ScopeID   int64
	Hash      []byte
	ExpiresAt int64
}

// Create inserts a new token record within the given transaction.
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, p CreateParams) error {
	_, err := r.tokens.WithTx(tx).Insert(ctx, gentokens.InsertParams{
		AccountID:    p.AccountID,
		TokenScopeID: p.ScopeID,
		Hash:         p.Hash,
		ExpiresAt:    p.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("tokens.Repository.Create: %w", err)
	}
	return nil
}

// Delete removes all tokens for accountID within scopeID.
func (r *Repository) Delete(ctx context.Context, tx *sql.Tx, accountID string, scopeID int64) error {
	if err := r.tokens.WithTx(tx).DeleteForAccountAndScope(ctx, gentokens.DeleteForAccountAndScopeParams{
		AccountID:    accountID,
		TokenScopeID: scopeID,
	}); err != nil {
		return fmt.Errorf("tokens.Repository.Delete: %w", err)
	}
	return nil
}

// VerifyAndIncrement validates the token hash and increments its attempt counter.
func (r *Repository) VerifyAndIncrement(ctx context.Context, tx *sql.Tx, hash []byte) (Record, error) {
	row, err := r.tokens.WithTx(tx).VerifyAndIncrement(ctx, hash)
	if err != nil {
		return Record{}, fmt.Errorf("tokens.Repository.VerifyAndIncrement: %w", err)
	}
	return Record{
		AccountID:    row.AccountID,
		TokenScopeID: row.TokenScopeID,
	}, nil
}
