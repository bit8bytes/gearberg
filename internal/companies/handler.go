// Package companies handles company-related HTTP routes, business logic, and data access.
package companies

import (
	"context"
	"errors"
	"fmt"
	"html/template"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
	"github.com/tobiasgleiter/forma"
)

// Handler holds dependencies for company HTTP handlers.
type Handler struct {
	svc   *Service
	cache map[string]*template.Template
}

// NewHandler returns a new Handler wired with svc.
func NewHandler(svc *Service, cache map[string]*template.Template) *Handler {
	return &Handler{svc: svc, cache: cache}
}

// Routes registers company routes on m.
func (h *Handler) Routes(m *forma.HTML) {
	forma.Get(m, forma.Operation{
		Path:     "/companies/{$}",
		Template: h.cache[pages.Companies.File],
	}, h.listCompanies)

	forma.Get(m, forma.Operation{
		Path:     "/companies/new",
		Template: h.cache[pages.CompaniesNew.File],
	}, h.getCompanyNew)

	forma.Post(m, forma.Operation{
		Path:        "/companies/new",
		Template:    h.cache[pages.CompaniesNew.File],
		RedirectURL: "/companies",
	}, h.createCompany)

	forma.Get(m, forma.Operation{
		Path:     "/companies/{company_id}",
		Template: h.cache[pages.CompanySettingsCompany.File],
	}, h.getSettingsCompany)

	forma.Post(m, forma.Operation{
		Path:     "/companies/{company_id}",
		Template: h.cache[pages.CompanySettingsCompany.File],
	}, h.updateCompany)

	forma.Post(m, forma.Operation{
		Path:        "/companies/{company_id}/delete",
		Template:    h.cache[pages.CompanySettingsCompany.File],
		RedirectURL: "/companies",
	}, h.deleteCompanyForm)
}

type deleteCompanyFormInput struct {
	CompanyID string `path:"company_id"`
}

type deleteCompanyFormOutput struct{}

func (h *Handler) deleteCompanyForm(ctx context.Context, in *deleteCompanyFormInput) (*deleteCompanyFormOutput, error) {
	if err := h.svc.Delete(ctx, in.CompanyID); err != nil {
		return nil, fmt.Errorf("deleteCompanyForm: %w", err)
	}
	return &deleteCompanyFormOutput{}, nil
}

type listCompaniesInput struct{}

type listCompaniesOutput struct {
	Companies    []Company
	MaxCompanies int
}

func (h *Handler) listCompanies(ctx context.Context, _ *listCompaniesInput) (*listCompaniesOutput, error) {
	companies, err := h.svc.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listCompanies: %w", err)
	}
	return &listCompaniesOutput{Companies: companies, MaxCompanies: h.svc.MaxCompanies()}, nil
}

type createCompanyInput struct {
	Name string `form:"name" required:"true" maxLength:"100"`
}

type companyFormOutput struct{}

func (h *Handler) getCompanyNew(_ context.Context, _ *createCompanyInput) (*companyFormOutput, error) {
	return &companyFormOutput{}, nil
}

func (h *Handler) createCompany(ctx context.Context, in *createCompanyInput) (*companyFormOutput, error) {
	_, err := h.svc.Create(ctx, CreateCompany{
		ID:   ksuid.New().String(),
		Name: in.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			return nil, &forma.ValidationError{Field: map[string]string{"name": "A company with this name already exists."}}
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := h.svc.MaxCompanies()
			return nil, &forma.ValidationError{Field: map[string]string{"name": fmt.Sprintf("Company limit reached. Only %d company allowed.", limit)}}
		}
		return nil, fmt.Errorf("createCompany: %w", err)
	}
	return &companyFormOutput{}, nil
}

type settingsCompanyInput struct {
	CompanyID string `path:"company_id"`
	Name      string `form:"name" required:"true" maxLength:"100"`
}

type settingsCompanyOutput struct {
	Company *Company
}

func (h *Handler) getSettingsCompany(ctx context.Context, in *settingsCompanyInput) (*settingsCompanyOutput, error) {
	company, err := h.svc.GetByID(ctx, in.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("getSettingsCompany: %w", err)
	}
	return &settingsCompanyOutput{Company: company}, nil
}

func (h *Handler) updateCompany(ctx context.Context, in *settingsCompanyInput) (*settingsCompanyOutput, error) {
	company, err := h.svc.Update(ctx, UpdateCompany{
		ID:   in.CompanyID,
		Name: in.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			return nil, &forma.ValidationError{Field: map[string]string{"name": "A company with this name already exists."}}
		}
		return nil, fmt.Errorf("updateCompany: %w", err)
	}
	return &settingsCompanyOutput{Company: company}, nil
}
