package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type orgSettingsData struct {
	Settings *settings.OrgSettings
	OrgID    string
}

func (app *application) getOrgSettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	s, err := app.services.orgsettings.GetByOrgID(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve org settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &settings.Form{}
	data.Data = orgSettingsData{Settings: s, OrgID: id}
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

	s, err := app.services.orgsettings.Upsert(ctx, settings.UpsertOrgSettings{
		ID:       ksuid.New().String(),
		OrgID:    id,
		Currency: form.Currency,
		VatRate:  form.VatRateBasisPoints(),
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
	data.Form = &settings.Form{}
	data.Data = orgSettingsData{Settings: s, OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettings, data)
}
