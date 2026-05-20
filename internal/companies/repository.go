package companies

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bit8bytes/gearberg/database"
	gencompanies "github.com/bit8bytes/gearberg/database/queries/gen/companies"
)

// Repository provides data access for companies.
type Repository struct {
	db        *sql.DB
	companies gencompanies.Querier
}

// NewRepository returns a new Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:        db,
		companies: gencompanies.New(db),
	}
}

// Count returns the total number of companies.
func (r *Repository) Count(ctx context.Context) (int64, error) {
	n, err := r.companies.CountCompanies(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count companies: %w", err)
	}
	return n, nil
}

// GetAll returns all companies.
func (r *Repository) GetAll(ctx context.Context) ([]Company, error) {
	companies, err := r.companies.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all companies: %w", err)
	}

	result := make([]Company, len(companies))
	for i, c := range companies {
		result[i] = Company{
			ID:   c.ID,
			Name: c.Name,
		}
	}
	return result, nil
}

// GetByID returns the company with the given ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*Company, error) {
	c, err := r.companies.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get company by id: %w", err)
	}
	return &Company{
		ID:   c.ID,
		Name: c.Name,
	}, nil
}

// CreateCompany holds the data required to create a new company.
type CreateCompany struct {
	ID   string
	Name string
}

// UpdateCompany holds the data required to update a company.
type UpdateCompany struct {
	ID   string
	Name string
}

// Update updates the name of the company with the given ID.
func (r *Repository) Update(ctx context.Context, u UpdateCompany) (*Company, error) {
	company, err := r.companies.UpdateByID(ctx, gencompanies.UpdateByIDParams{
		Name: u.Name,
		ID:   u.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update company: %w", database.NormalizeError(err))
	}
	return &Company{ID: company.ID, Name: company.Name}, nil
}

// Delete removes the company with the given ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := r.companies.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete company: %w", err)
	}
	return nil
}

// Create creates a new company.
func (r *Repository) Create(ctx context.Context, createCompany CreateCompany) (*Company, error) {
	company, err := r.companies.Create(ctx, gencompanies.CreateParams{
		ID:   createCompany.ID,
		Name: createCompany.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", database.NormalizeError(err))
	}

	return &Company{
		ID:   company.ID,
		Name: company.Name,
	}, nil
}
