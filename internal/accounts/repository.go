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

// Create inserts a new account and returns its ID.
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, id, email string) (string, error) {
	row, err := r.accounts.WithTx(tx).Create(ctx, genaccounts.CreateParams{
		ID:    id,
		Email: email,
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
