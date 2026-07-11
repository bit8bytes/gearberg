package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/segmentio/ksuid"
)

type orgsData struct {
	Orgs    []orgs.Org
	MaxOrgs int
}

func (app *application) getOrgs(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()

	allOrgs, err := app.services.orgs.GetAll(ctx)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve orgs.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Data = orgsData{
		Orgs:    allOrgs,
		MaxOrgs: app.services.orgs.MaxOrgs(),
	}

	return app.html.Render(w, r, http.StatusOK, pages.Orgs, data)
}

func (app *application) getOrgsNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Form = &orgs.Form{}
	return app.html.Render(w, r, http.StatusOK, pages.OrgsNew, data)
}

func (app *application) postOrgsNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()

	form, err := orgs.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *orgs.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.OrgsNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	org, err := app.services.orgs.Create(ctx, orgs.CreateOrg{
		ID:   ksuid.New().String(),
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A org with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.orgs.MaxOrgs()
			form.AddError("name", fmt.Sprintf("Org limit reached. Only %d orgs allowed.", limit))
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to create org.",
			Code:    http.StatusInternalServerError,
		}
	}

	_, err = app.services.orgsettings.Upsert(ctx, settings.UpsertOrgSettings{
		ID:       ksuid.New().String(),
		OrgID:    org.ID,
		Currency: app.options.DefaultCurrency,
		VatRate:  app.options.DefaultVatRateBasisPoints(),
		Timezone: app.options.DefaultTimezone,
	})
	if err != nil {
		return &httperr.Error{
			Error:   fmt.Errorf("postOrgsNew: seed default settings: %w", err),
			Message: "Failed to initialise org settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/orgs/%s/equipment", org.ID), http.StatusSeeOther)
	return nil
}

type settingsOrgData struct {
	Org          *orgs.Org
	OrgID        string
	StorageUsage storage.Usage
}

func (app *application) getSettingsOrg(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	org, err := app.services.orgs.GetByID(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve org.",
			Code:    http.StatusInternalServerError,
		}
	}

	usage, err := app.services.storageManager.Info(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve storage info.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &orgs.Form{}
	data.Data = settingsOrgData{Org: org, OrgID: id, StorageUsage: usage}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettingsDetails, data)
}

func (app *application) postSettingsOrg(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := orgs.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *orgs.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = settingsOrgData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.OrgSettingsDetails, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	org, err := app.services.orgs.Update(ctx, orgs.UpdateOrg{
		ID:   id,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A org with this name already exists.")
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to update org.",
			Code:    http.StatusInternalServerError,
		}
	}

	usage, err := app.services.storageManager.Info(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve storage info.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &orgs.Form{}
	data.Data = settingsOrgData{Org: org, OrgID: id, StorageUsage: usage}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettingsDetails, data)
}

func (app *application) postDeleteOrg(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	if err := app.services.orgs.Delete(ctx, id); err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete org.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/orgs", http.StatusSeeOther)
	return nil
}
