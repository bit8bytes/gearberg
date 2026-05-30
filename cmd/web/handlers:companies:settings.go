package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/internal/companies/settings"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type companySettingsData struct {
	Settings  *settings.CompanySettings
	CompanyID string
}

func (app *application) getCompanySettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("company_id")

	s, err := app.services.companysettings.GetByCompanyID(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve company settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &settings.Form{}
	data.Data = companySettingsData{Settings: s, CompanyID: id}
	return app.html.Render(w, r, http.StatusOK, pages.CompanySettingsConfig, data)
}

func (app *application) postCompanySettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("company_id")

	form, err := settings.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *settings.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = companySettingsData{CompanyID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.CompanySettingsConfig, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	s, err := app.services.companysettings.Upsert(ctx, settings.UpsertCompanySettings{
		ID:        ksuid.New().String(),
		CompanyID: id,
		Currency:  form.Currency,
		VatRate:   form.VatRate,
		Timezone:  form.Timezone,
	})
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to save company settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &settings.Form{}
	data.Data = companySettingsData{Settings: s, CompanyID: id}
	return app.html.Render(w, r, http.StatusOK, pages.CompanySettingsConfig, data)
}
