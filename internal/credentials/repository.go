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

// Package credentials manages credential records used for authentication.
package credentials

import (
	"context"
	"database/sql"
	"fmt"

	gencredentials "github.com/bit8bytes/gearberg/internal/database/queries/gen/credentials"
)

// Record holds the data returned by GetByEmail for authentication.
type Record struct {
	AccountID  string
	SecretData []byte
}

// Repository provides data access for credentials.
type Repository struct {
	credentials *gencredentials.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db gencredentials.DBTX) *Repository {
	return &Repository{
		credentials: gencredentials.New(db),
	}
}

// Create inserts a new credential record for accountID within the given transaction.
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, accountID string, typeID int64, secretData []byte) error {
	if err := r.credentials.WithTx(tx).Create(ctx, gencredentials.CreateParams{
		AccountID:  accountID,
		TypeID:     typeID,
		SecretData: secretData,
	}); err != nil {
		return fmt.Errorf("credentials.Repository.Create: %w", err)
	}
	return nil
}

// GetByEmail looks up a credential by account email and credential type.
func (r *Repository) GetByEmail(ctx context.Context, email string, typeID int64) (Record, error) {
	row, err := r.credentials.GetByEmail(ctx, gencredentials.GetByEmailParams{
		Email:  email,
		TypeID: typeID,
	})
	if err != nil {
		return Record{}, fmt.Errorf("credentials.Repository.GetByEmail: %w", err)
	}
	return Record{
		AccountID:  row.AccountID,
		SecretData: row.SecretData,
	}, nil
}

// GetByAccountID looks up a credential by account ID and credential type.
func (r *Repository) GetByAccountID(ctx context.Context, accountID string, typeID int64) (Record, error) {
	row, err := r.credentials.GetByAccountID(ctx, gencredentials.GetByAccountIDParams{
		AccountID: accountID,
		TypeID:    typeID,
	})
	if err != nil {
		return Record{}, fmt.Errorf("credentials.Repository.GetByAccountID: %w", err)
	}
	return Record{
		AccountID:  row.AccountID,
		SecretData: row.SecretData,
	}, nil
}

// UpdatePassword replaces the hashed password for accountID within the given transaction.
func (r *Repository) UpdatePassword(ctx context.Context, tx *sql.Tx, accountID string, hash string) error {
	if err := r.credentials.WithTx(tx).UpdateSecret(ctx, gencredentials.UpdateSecretParams{
		AccountID:  accountID,
		TypeID:     int64(PasswordType),
		SecretData: []byte(hash),
	}); err != nil {
		return fmt.Errorf("credentials.Repository.UpdatePassword: %w", err)
	}
	return nil
}
