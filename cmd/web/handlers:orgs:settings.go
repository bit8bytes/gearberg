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
	"net/http"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs/settings"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

type orgSettingsData struct {
	OrgID string
}

func (app *application) getOrgSettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	s, err := app.services.orgsettings.Get(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	f := s.Form()
	data.Form = &f
	data.Data = orgSettingsData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettings, data)
}

func (app *application) postOrgSettings(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := settings.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	fail := func(f *settings.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = orgSettingsData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.OrgSettings, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	s, err := app.services.orgsettings.Update(ctx, settings.UpdateOrgSettings{
		OrgID:    id,
		Currency: form.Currency,
		VatRate:  form.VatRate,
		Timezone: form.Timezone,
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	f := s.Form()
	data.Form = &f
	data.Data = orgSettingsData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.OrgSettings, data)
}
