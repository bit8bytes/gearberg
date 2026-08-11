package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs"
	"github.com/bit8bytes/gearberg/internal/sessions"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/internal/uid"
	"github.com/bit8bytes/gearberg/pkg/htmx"
)

type orgsData struct {
	Orgs []orgs.Org
	Max  int
}

func (app *application) getOrgs(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	session := sessions.MustFromRequest(r)

	allOrgs, err := app.services.orgs.List(ctx, session.AccountID)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve orgs.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Data = orgsData{
		Orgs: allOrgs,
		Max:  app.options.MaxOrgs,
	}

	return app.html.Render(w, r, http.StatusOK, pages.SettingsOrganizations, data)
}

func (app *application) getOrgsNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Form = &orgs.Form{}
	return app.html.Render(w, r, http.StatusOK, pages.OrgsNew, data)
}

type settingsOrgData struct {
	OrgID        string
	StorageUsage storage.Usage
}

func (app *application) getSettingsOrg(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	org, err := app.services.orgs.Get(ctx, id)
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
	data.Form = &orgs.Form{DisplayName: org.DisplayName}
	data.Data = settingsOrgData{OrgID: id, StorageUsage: usage}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettingsDetails, data)
}

func (app *application) deleteOrg(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	session := sessions.MustFromRequest(r)

	if err := app.services.orgs.Delete(ctx, orgID, session.AccountID); err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete org. You must be the sole owner to delete it.",
			Code:    http.StatusForbidden,
		}
	}

	if htmx.IsRequest(r) {
		htmx.Redirect(w, r, "/orgs", http.StatusOK)
	} else {
		http.Redirect(w, r, "/orgs", http.StatusSeeOther)
	}
	return nil
}

func (app *application) postSettingsOrg(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id, err := uid.Parse(r.PathValue("org_id"))
	if err != nil {
		return &httperr.Error{Error: err, Message: "Invalid org ID.", Code: http.StatusBadRequest}
	}

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

	err = app.services.orgs.Update(ctx, orgs.UpdateParams{
		ID:          id,
		DisplayName: form.DisplayName,
	})
	if err != nil {
		if errors.Is(err, orgs.ErrConflict) {
			form.AddError("name", "A org with this name already exists.")
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to update org.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/orgs/"+id, http.StatusSeeOther) //nolint:gosec // id is a parsed and validated KSUID, not an open redirect
	return nil
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

	session := sessions.MustFromRequest(r)
	orgID, err := app.services.orgs.Create(ctx, orgs.CreateOrg{
		ID:          uid.New(),
		DisplayName: form.DisplayName,
		AccountID:   session.AccountID,
	})
	if err != nil {
		if errors.Is(err, orgs.ErrConflict) {
			form.AddError("name", "A org with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, orgs.ErrLimitExceeded) {
			form.AddError("name", fmt.Sprintf("Org limit reached. Only %d orgs allowed.", app.options.MaxOrgs))
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to create org.",
			Code:    http.StatusInternalServerError,
		}
	}

	_, err = app.services.orgsettings.Create(ctx, uid.New(), orgID)
	if err != nil {
		return &httperr.Error{
			Error:   fmt.Errorf("postOrgsNew: seed default settings: %w", err),
			Message: "Failed to initialise org settings.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/orgs/%s/equipment", orgID), http.StatusSeeOther)
	return nil
}
