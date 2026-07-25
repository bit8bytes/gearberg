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
