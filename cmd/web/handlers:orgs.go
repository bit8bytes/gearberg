// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
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
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Data = orgsData{
		Orgs: allOrgs,
		Max:  app.options.Limits.MaxOrgs,
	}

	return app.html.Render(w, r, http.StatusOK, pages.SettingsOrganizations, data)
}

func (app *application) getOrgPicker(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	session := sessions.MustFromRequest(r)

	allOrgs, err := app.services.orgs.List(ctx, session.AccountID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Data = orgsData{
		Orgs: allOrgs,
		Max:  app.options.Limits.MaxOrgs,
	}

	return app.html.Render(w, r, http.StatusOK, pages.OrgPicker, data)
}

func (app *application) getOrgsNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Form = orgs.NewForm()
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
		return httperr.InternalServerError(err)
	}

	usage, err := app.services.storageManager.Info(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Form = &orgs.Form{DisplayName: org.DisplayName}
	data.Data = settingsOrgData{
		OrgID:        id,
		StorageUsage: usage,
	}
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
		return httperr.BadRequest(err)
	}

	form, err := orgs.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	fail := func(f *orgs.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = settingsOrgData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.OrgSettingsDetails, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	err = app.services.orgs.Update(ctx, orgs.UpdateParams{
		ID:          id,
		DisplayName: form.DisplayName,
	})
	if err != nil {
		if errors.Is(err, orgs.ErrConflict) {
			form.AddError("name", "A org with this name already exists.")
			return fail(&form)
		}
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+id, http.StatusSeeOther) //nolint:gosec // id is a parsed and validated KSUID, not an open redirect
	return nil
}

func (app *application) postOrgsNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()

	form, err := orgs.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	fail := func(f *orgs.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.OrgsNew, data)
	}

	if !form.Validate() {
		return fail(&form)
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
			return fail(&form)
		}
		if errors.Is(err, orgs.ErrLimitExceeded) {
			form.AddError("name", fmt.Sprintf("Org limit reached. Only %d orgs allowed.", app.options.Limits.MaxOrgs))
			return fail(&form)
		}
		return httperr.InternalServerError(err)
	}

	_, err = app.services.orgsettings.Create(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(fmt.Errorf("postOrgsNew: seed default settings: %w", err))
	}

	http.Redirect(w, r, fmt.Sprintf("/orgs/%s/equipment", orgID), http.StatusSeeOther)
	return nil
}
