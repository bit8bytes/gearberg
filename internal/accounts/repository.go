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

// Package accounts provides accounts functionality.
package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bit8bytes/gearberg/internal/database"
	genaccounts "github.com/bit8bytes/gearberg/internal/database/queries/gen/accounts"
)

// Record holds the data returned by account lookups.
type Record struct {
	ID            string
	Email         string
	EmailVerified *time.Time
}

// Repository provides data access for accounts.
type Repository struct {
	accounts *genaccounts.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db genaccounts.DBTX) *Repository {
	return &Repository{
		accounts: genaccounts.New(db),
	}
}

// Create inserts a new account and returns its ID. emailVerifiedAt is nil for
// password-based sign-up (not yet verified) and set to the current time for
// OIDC accounts where the provider already confirmed the email.
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, id, email string, emailVerifiedAt *int64) (string, error) {
	row, err := r.accounts.WithTx(tx).Create(ctx, genaccounts.CreateParams{
		ID:            id,
		Email:         email,
		EmailVerified: database.NullInt64(emailVerifiedAt),
	})
	if err != nil {
		if errors.Is(database.NormalizeError(err), database.ErrUniqueConstraint) {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("accounts.Repository.Create: %w", err)
	}
	return row.ID, nil
}

// Delete removes an account by ID.
func (r *Repository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	if err := r.accounts.WithTx(tx).Delete(ctx, id); err != nil {
		return fmt.Errorf("accounts.Repository.Delete: %w", err)
	}
	return nil
}

// GetByAccountID retrieves an account by its ID.
func (r *Repository) GetByAccountID(ctx context.Context, accountID string) (Record, error) {
	row, err := r.accounts.GetByAccountID(ctx, accountID)
	if err != nil {
		return Record{}, fmt.Errorf("accounts.Repository.GetByAccountID: %w", err)
	}
	var emailVerified *time.Time
	if row.EmailVerified.Valid {
		t := time.Unix(row.EmailVerified.Int64, 0)
		emailVerified = &t
	}
	return Record{
		ID:            row.ID,
		Email:         row.Email,
		EmailVerified: emailVerified,
	}, nil
}
