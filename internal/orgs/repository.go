package orgs

import (
	"context"
	"database/sql"
	"fmt"

	genorgs "github.com/bit8bytes/gearberg/internal/database/queries/gen/orgs"
)

// Repository provides data access for organizations and their members.
type Repository struct {
	orgs    *genorgs.Queries
	maxOrgs int64
}

// NewRepository returns a new Repository.
func NewRepository(db genorgs.DBTX, maxOrgs int64) *Repository {
	if maxOrgs <= 0 {
		panic("orgs.NewRepository: MaxOrgs must be greater than zero")
	}
	return &Repository{
		orgs:    genorgs.New(db),
		maxOrgs: maxOrgs,
	}
}

// Create inserts a new organization and returns its ID.
func (r *Repository) Create(ctx context.Context, tx *sql.Tx, accountID, id, displayName string) (string, error) {
	row, err := r.orgs.WithTx(tx).Create(ctx, genorgs.CreateParams{
		ID:          id,
		DisplayName: displayName,
		AccountID:   accountID,
		MaxOrgs:     r.maxOrgs,
	})
	if err != nil {
		return "", fmt.Errorf("orgs.Repository.Create: %w", err)
	}
	return row.ID, nil
}

// CreateMember adds an account as a member of an organization with the given role.
func (r *Repository) CreateMember(ctx context.Context, tx *sql.Tx, accountID, organizationID string, roleID int64) error {
	if err := r.orgs.WithTx(tx).CreateMember(ctx, genorgs.CreateMemberParams{
		AccountID: accountID,
		OrgID:     organizationID,
		RoleID:    roleID,
	}); err != nil {
		return fmt.Errorf("orgs.Repository.CreateMember: %w", err)
	}
	return nil
}

// Max returns the configured maximum number of organizations.
func (r *Repository) Max() int {
	return int(r.maxOrgs)
}

// List returns organizations the given account is a member of, ordered by creation time.
func (r *Repository) List(ctx context.Context, accountID string) ([]Org, error) {
	rows, err := r.orgs.List(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("orgs.Repository.List: %w", err)
	}
	orgs := make([]Org, len(rows))
	for i, row := range rows {
		orgs[i] = Org{
			ID:          row.ID,
			DisplayName: row.DisplayName,
		}
	}
	return orgs, nil
}

// GetFirstByAccountID returns the ID of the first organization the account belongs to.
func (r *Repository) GetFirstByAccountID(ctx context.Context, accountID string) (string, error) {
	id, err := r.orgs.GetFirstByAccountID(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("orgs.Repository.GetFirstByAccountID: %w", err)
	}
	return id, nil
}

// DeleteOwnedByAccount deletes all organizations where accountID is the sole owner.
func (r *Repository) DeleteOwnedByAccount(ctx context.Context, tx *sql.Tx, accountID string) error {
	if err := r.orgs.WithTx(tx).DeleteOwnedByAccountID(ctx, accountID); err != nil {
		return fmt.Errorf("orgs.Repository.DeleteOwnedByAccount: %w", err)
	}
	return nil
}

// Delete removes a single org if accountID is its sole owner.
func (r *Repository) Delete(ctx context.Context, orgID, accountID string) error {
	if err := r.orgs.Delete(ctx, genorgs.DeleteParams{OrgID: orgID, AccountID: accountID}); err != nil {
		return fmt.Errorf("orgs.Repository.Delete: %w", err)
	}
	return nil
}

// GetMember returns an error if accountID is not a member of orgID.
func (r *Repository) GetMember(ctx context.Context, orgID, accountID string) error {
	if _, err := r.orgs.GetMemberByAccountIDAndOrgID(ctx, genorgs.GetMemberByAccountIDAndOrgIDParams{
		AccountID: accountID,
		OrgID:     orgID,
	}); err != nil {
		return fmt.Errorf("orgs.Repository.GetMember: %w", err)
	}
	return nil
}

// UpdateParams holds the fields to update on an organization.
type UpdateParams struct {
	ID               string
	DisplayName      string
	PhoneCountryCode string
	PhoneNumber      string
	Email            string
}

// Get retrieves the organization with the given ID.
func (r *Repository) Get(ctx context.Context, id string) (Org, error) {
	row, err := r.orgs.Get(ctx, id)
	if err != nil {
		return Org{}, fmt.Errorf("orgs.Repository.Get: %w", err)
	}
	return Org{
		ID:          row.ID,
		DisplayName: row.DisplayName,
	}, nil
}

// Update saves display_name and contact fields for the given organization.
func (r *Repository) Update(ctx context.Context, p UpdateParams) error {
	err := r.orgs.Update(ctx, genorgs.UpdateParams{
		ID:          p.ID,
		DisplayName: p.DisplayName,
	})
	if err != nil {
		return fmt.Errorf("orgs.Repository.Update: %w", err)
	}
	return nil
}
