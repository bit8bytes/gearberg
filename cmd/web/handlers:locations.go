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
	"net/url"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/locations"
	"github.com/bit8bytes/gearberg/internal/templates/fragments"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/pkg/htmx"
)

type locationsData struct {
	OrgID        string
	Locations    []locations.Location
	MaxLocations int
	SelectedID   string
	SelectedName string
}

type locationData struct {
	OrgID string
	ID    string
}

func (app *application) getLocations(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	locs, err := app.services.locations.List(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Form = locations.NewForm()
	data.Data = locationsData{
		OrgID:        id,
		Locations:    locs,
		MaxLocations: app.options.Limits.MaxOrgLocations,
	}
	return app.html.Render(w, r, http.StatusOK, pages.LocationsIndex, data)
}

func (app *application) getLocationNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")
	data := app.html.TemplateData(r)
	data.Form = locations.NewForm()
	data.Data = locationData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.LocationsNew, data)
}

func (app *application) postLocationNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := locations.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	fail := func(f *locations.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = locationData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.LocationsNew, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	_, err = app.services.locations.Create(ctx, locations.CreateLocation{
		OrgID: id,
		Name:  form.Name,
	})
	if err != nil {
		if errors.Is(err, locations.ErrConflict) {
			form.AddError("name", "A location with this name already exists.")
			return fail(&form)
		}
		if errors.Is(err, locations.ErrLimitExceeded) {
			limit := app.options.Limits.MaxOrgLocations
			form.AddError("name", fmt.Sprintf("Location limit reached. Only %d locations allowed per org.", limit))
			return fail(&form)
		}
		return httperr.InternalServerError(err)
	}

	dest := "/orgs/" + url.PathEscape(id) + "/settings/locations"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is a hard-coded path with org_id from path value.
	return nil
}

func (app *application) getLocation(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	locID := r.PathValue("id")

	loc, err := app.services.locations.GetByID(ctx, locID)
	if err != nil {
		if errors.Is(err, locations.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Form = &locations.Form{Name: loc.Name}
	data.Data = locationData{
		OrgID: orgID,
		ID:    locID,
	}
	return app.html.Render(w, r, http.StatusOK, pages.LocationsDetail, data)
}

func (app *application) postLocation(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	locID := r.PathValue("id")

	form, err := locations.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	loc, err := app.services.locations.GetByID(ctx, locID)
	if err != nil {
		if errors.Is(err, locations.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	fail := func(f *locations.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = locationData{
			OrgID: loc.OrgID,
			ID:    locID,
		}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.LocationsDetail, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	_, err = app.services.locations.Update(ctx, locations.UpdateLocation{
		ID:   locID,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, locations.ErrConflict) {
			form.AddError("name", "A location with this name already exists.")
			return fail(&form)
		}
		return httperr.InternalServerError(err)
	}

	dest := "/orgs/" + loc.OrgID + "/settings/locations/" + loc.ID
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is a hard-coded path with validated IDs.
	return nil
}

func (app *application) postDeleteLocation(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	locID := r.PathValue("id")

	if err := app.services.locations.Delete(ctx, locID); err != nil {
		return httperr.InternalServerError(err)
	}

	dest := "/orgs/" + url.PathEscape(orgID) + "/settings/locations"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is a hard-coded path with org_id from path value.
	return nil
}

func (app *application) getEquipmentLocationsFragment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/settings/locations", http.StatusSeeOther)
		return nil
	}

	selected := r.URL.Query().Get("selected")
	selectedName := r.URL.Query().Get("selected_name")

	locs, err := app.services.locations.List(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(fmt.Errorf("getEquipmentLocationsFragment: %w", err))
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = locationsData{
		OrgID:        orgID,
		Locations:    locs,
		SelectedID:   selected,
		SelectedName: selectedName,
	}
	return app.html.RenderFragment(w, r, http.StatusOK, fragments.WarehouseLocations, tmplData)
}
