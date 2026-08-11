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
package federated

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bit8bytes/gearberg/internal/database"
	genfederated "github.com/bit8bytes/gearberg/internal/database/queries/gen/federated"
)

// Repository provides data access for federated identity records.
type Repository struct {
	q *genfederated.Queries
}

// NewRepository returns a new Repository.
func NewRepository(db genfederated.DBTX) *Repository {
	return &Repository{q: genfederated.New(db)}
}

// GetAccountIDByProviderSubject looks up the account ID for a given provider + subject pair.
// Returns database.ErrNotFound when no matching identity exists.
func (r *Repository) GetAccountIDByProviderSubject(ctx context.Context, providerID int64, subject string) (string, error) {
	accountID, err := r.q.GetAccountIDByProviderSubject(ctx, genfederated.GetAccountIDByProviderSubjectParams{
		ProviderID:      providerID,
		ProviderSubject: subject,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", database.ErrNotFound
		}
		return "", fmt.Errorf("federated.Repository.GetAccountIDByProviderSubject: %w", err)
	}
	return accountID, nil
}

// Create inserts a new federated identity record within the given transaction.
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, accountID string, providerID int64, subject string) error {
	if err := r.q.WithTx(tx).Create(ctx, genfederated.CreateParams{
		AccountID:       accountID,
		ProviderID:      providerID,
		ProviderSubject: subject,
	}); err != nil {
		return fmt.Errorf("federated.Repository.Create: %w", err)
	}
	return nil
}
