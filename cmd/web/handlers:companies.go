package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/internal/companies"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type companiesData struct {
	Companies    []companies.Company
	MaxCompanies int
}

func (app *application) getCompanies(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()

	allCompanies, err := app.services.companies.GetAll(ctx)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve companies.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Data = companiesData{
		Companies:    allCompanies,
		MaxCompanies: app.services.companies.MaxCompanies(),
	}

	return app.render(w, r, http.StatusOK, pages.Companies, data)
}

func (app *application) getCompaniesNew(w http.ResponseWriter, r *http.Request) *appError {
	data := app.newTemplateData(r)
	data.Form = &companies.Form{}
	return app.render(w, r, http.StatusOK, pages.CompaniesNew, data)
}

func (app *application) postCompaniesNew(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()

	form, err := companies.Parse(r)
	if err != nil {
		return &appError{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *companies.Form) *appError {
		data := app.newTemplateData(r)
		data.Form = f
		return app.render(w, r, http.StatusUnprocessableEntity, pages.CompaniesNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	company, err := app.services.companies.Create(ctx, companies.CreateCompany{
		ID:   ksuid.New().String(),
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A company with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.companies.MaxCompanies()
			form.AddError("name", fmt.Sprintf("Company limit reached. Only %d company allowed.", limit))
			return reRender(&form)
		}
		return &appError{
			Error:   err,
			Message: "Failed to create company.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/companies/%s/inventory", company.ID), http.StatusSeeOther)
	return nil
}

type settingsCompanyData struct {
	Company      *companies.Company
	CompanyID    string
	StorageUsage storage.Usage
}

func (app *application) getSettingsCompany(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	company, err := app.services.companies.GetByID(ctx, id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve company.",
			Code:    http.StatusInternalServerError,
		}
	}

	usage, err := app.services.storageManager.Info(ctx, id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve storage info.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Form = &companies.Form{}
	data.Data = settingsCompanyData{Company: company, CompanyID: id, StorageUsage: usage}
	return app.render(w, r, http.StatusOK, pages.CompanySettingsCompany, data)
}

func (app *application) postSettingsCompany(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	form, err := companies.Parse(r)
	if err != nil {
		return &appError{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *companies.Form) *appError {
		data := app.newTemplateData(r)
		data.Form = f
		data.Data = settingsCompanyData{CompanyID: id}
		return app.render(w, r, http.StatusUnprocessableEntity, pages.CompanySettingsCompany, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	company, err := app.services.companies.Update(ctx, companies.UpdateCompany{
		ID:   id,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A company with this name already exists.")
			return reRender(&form)
		}
		return &appError{
			Error:   err,
			Message: "Failed to update company.",
			Code:    http.StatusInternalServerError,
		}
	}

	usage, err := app.services.storageManager.Info(ctx, id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve storage info.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Form = &companies.Form{}
	data.Data = settingsCompanyData{Company: company, CompanyID: id, StorageUsage: usage}
	return app.render(w, r, http.StatusOK, pages.CompanySettingsCompany, data)
}

func (app *application) postDeleteCompany(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	if err := app.services.companies.Delete(ctx, id); err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to delete company.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/companies", http.StatusSeeOther)
	return nil
}
