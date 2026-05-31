package orgs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	genorgs "github.com/bit8bytes/gearberg/database/queries/gen/orgs"
)

// Repository provides data access for orgs.
type Repository struct {
	db   *sql.DB
	orgs genorgs.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:   db,
		orgs: genorgs.New(db),
	}
}

// Count returns the total number of orgs.
func (r *Repository) Count(ctx context.Context) (int64, error) {
	n, err := r.orgs.CountOrgs(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count orgs: %w", err)
	}
	return n, nil
}

// GetAll returns all orgs.
func (r *Repository) GetAll(ctx context.Context) ([]Org, error) {
	orgs, err := r.orgs.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orgs: %w", err)
	}

	result := make([]Org, len(orgs))
	for i, c := range orgs {
		result[i] = Org{
			ID:   c.ID,
			Name: c.Name,
		}
	}
	return result, nil
}

// GetByID returns the org with the given ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*Org, error) {
	c, err := r.orgs.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get org by id: %w", err)
	}
	return &Org{
		ID:   c.ID,
		Name: c.Name,
	}, nil
}

// CreateOrg holds the data required to create a new org.
type CreateOrg struct {
	ID   string
	Name string
}

// UpdateOrg holds the data required to update a org.
type UpdateOrg struct {
	ID   string
	Name string
}

// Update updates the name of the org with the given ID.
func (r *Repository) Update(ctx context.Context, u UpdateOrg) (*Org, error) {
	org, err := r.orgs.UpdateByID(ctx, genorgs.UpdateByIDParams{
		Name: u.Name,
		ID:   u.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update org: %w", database.NormalizeError(err))
	}
	return &Org{ID: org.ID, Name: org.Name}, nil
}

// Delete removes the org with the given ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.orgs.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete org: %w", err)
	}
	return nil
}

// Create creates a new org.
func (r *Repository) Create(ctx context.Context, createOrg CreateOrg) (*Org, error) {
	org, err := r.orgs.Create(ctx, genorgs.CreateParams{
		ID:   createOrg.ID,
		Name: createOrg.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create org: %w", database.NormalizeError(err))
	}

	return &Org{
		ID:   org.ID,
		Name: org.Name,
	}, nil
}
