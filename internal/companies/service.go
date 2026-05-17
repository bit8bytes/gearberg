package companies

import (
	"context"
	"fmt"
)

// Service implements business logic for companies.
type Service struct {
	repo *Repository
}

// NewService returns a new Service backed by repo.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetAll returns all companies.
func (s *Service) GetAll(ctx context.Context) ([]Company, error) {
	companies, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all companies: %w", err)
	}
	return companies, nil
}

// GetByID returns the company with the given ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Company, error) {
	company, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get company by id: %w", err)
	}
	return company, nil
}

// Update updates the name of the company with the given ID.
func (s *Service) Update(ctx context.Context, u UpdateCompany) (*Company, error) {
	company, err := s.repo.Update(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}
	return company, nil
}

// Delete removes the company with the given ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete company: %w", err)
	}
	return nil
}

// Create creates a new company.
func (s *Service) Create(ctx context.Context, createCompany CreateCompany) (*Company, error) {
	company, err := s.repo.Create(ctx, createCompany)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}
	return company, nil
}
