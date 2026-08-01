package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/internal/templates/fragments"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/pkg/htmx"
)

type orgSettingsData struct {
	OrgID string
}

type orgCurrencyData struct {
	Currency settings.Currency
}

func (app *application) getOrgCurrencyFragment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/settings", http.StatusSeeOther)
		return nil
	}

	s, err := app.services.orgsettings.Get(ctx, orgID)
	if err != nil {
		return &httperr.Error{
			Error:   fmt.Errorf("getOrgCurrencyFragment: %w", err),
			Message: "Failed to retrieve org settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	var currency settings.Currency
	if s != nil {
		currency = s.Currency
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = orgCurrencyData{Currency: currency}
	return app.html.RenderFragment(w, r, http.StatusOK, fragments.OrgCurrency, tmplData)
}

func (app *application) getOrgSettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	s, err := app.services.orgsettings.Get(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve org settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	f := settings.FormFromOrgSettings(s)
	data.Form = &f
	data.Data = orgSettingsData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettings, data)
}

func (app *application) postOrgSettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := settings.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *settings.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = orgSettingsData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.OrgSettings, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	s, err := app.services.orgsettings.Update(ctx, settings.UpdateOrgSettings{
		OrgID:    id,
		Currency: form.Currency,
		VatRate:  form.ParsedVatRate(),
		Timezone: form.Timezone,
	})
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to save org settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	f := settings.FormFromOrgSettings(s)
	data.Form = &f
	data.Data = orgSettingsData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettings, data)
}
