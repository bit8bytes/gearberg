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
	"github.com/bit8bytes/gearberg/internal/manufacturers"
	"github.com/bit8bytes/gearberg/internal/templates/fragments"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/pkg/htmx"
)

type manufacturersData struct {
	OrgID            string
	Manufacturers    []manufacturers.Manufacturer
	MaxManufacturers int
	SelectedID       string
	SelectedName     string
}

type manufacturerData struct {
	OrgID string
	ID    string
}

func (app *application) getManufacturers(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	mfrs, err := app.services.manufacturers.List(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Form = manufacturers.NewForm()
	data.Data = manufacturersData{
		OrgID:            id,
		Manufacturers:    mfrs,
		MaxManufacturers: app.options.Limits.MaxOrgManufacturers,
	}
	return app.html.Render(w, r, http.StatusOK, pages.ManufacturersIndex, data)
}

func (app *application) getManufacturerNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")
	data := app.html.TemplateData(r)
	data.Form = manufacturers.NewForm()
	data.Data = manufacturerData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.ManufacturersNew, data)
}

func (app *application) postManufacturerNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := manufacturers.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	fail := func(f *manufacturers.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = manufacturerData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ManufacturersNew, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	_, err = app.services.manufacturers.Create(ctx, manufacturers.CreateManufacturer{
		OrgID: id,
		Name:  form.Name,
	})
	if err != nil {
		if errors.Is(err, manufacturers.ErrConflict) {
			form.AddError("name", "A manufacturer with this name already exists.")
			return fail(&form)
		}
		if errors.Is(err, manufacturers.ErrLimitExceeded) {
			limit := app.options.Limits.MaxOrgManufacturers
			form.AddError("name", fmt.Sprintf("Manufacturer limit reached. Only %d manufacturers allowed per org.", limit))
			return fail(&form)
		}
		return httperr.InternalServerError(err)
	}

	dest := "/orgs/" + url.PathEscape(id) + "/settings/manufacturers"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec
	return nil
}

func (app *application) getManufacturer(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	mfrID := r.PathValue("id")

	manufacturer, err := app.services.manufacturers.Get(ctx, mfrID)
	if err != nil {
		if errors.Is(err, manufacturers.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Form = &manufacturers.Form{Name: manufacturer.Name}
	data.Data = manufacturerData{
		OrgID: orgID,
		ID:    mfrID,
	}
	return app.html.Render(w, r, http.StatusOK, pages.ManufacturersDetail, data)
}

func (app *application) postManufacturer(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	mfrID := r.PathValue("id")

	form, err := manufacturers.Parse(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	mfr, err := app.services.manufacturers.Get(ctx, mfrID)
	if err != nil {
		if errors.Is(err, manufacturers.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	fail := func(f *manufacturers.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = manufacturerData{OrgID: mfr.OrgID, ID: mfrID}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ManufacturersDetail, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	_, err = app.services.manufacturers.Update(ctx, manufacturers.UpdateManufacturer{
		ID:   mfrID,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, manufacturers.ErrConflict) {
			form.AddError("name", "A manufacturer with this name already exists.")
			return fail(&form)
		}
		return httperr.InternalServerError(err)
	}

	dest := "/orgs/" + mfr.OrgID + "/settings/manufacturers/" + mfr.ID
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec
	return nil
}

func (app *application) postDeleteManufacturer(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	mfrID := r.PathValue("id")

	if err := app.services.manufacturers.Delete(ctx, mfrID); err != nil {
		if !errors.Is(err, manufacturers.ErrInUse) {
			return httperr.InternalServerError(err)
		}
		manufacturer, fetchErr := app.services.manufacturers.Get(ctx, mfrID)
		if fetchErr != nil {
			if errors.Is(fetchErr, manufacturers.ErrNotFound) {
				return httperr.NotFound(err)
			}
			return httperr.InternalServerError(fetchErr)
		}
		f := &manufacturers.Form{Name: manufacturer.Name}
		f.AddError("delete", "Cannot delete: this manufacturer is assigned to one or more inventory items.")
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = manufacturerData{OrgID: orgID, ID: mfrID}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ManufacturersDetail, data)
	}

	dest := "/orgs/" + url.PathEscape(orgID) + "/settings/manufacturers"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec
	return nil
}

func (app *application) getEquipmentManufacturersFragment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/settings/manufacturers", http.StatusSeeOther)
		return nil
	}

	selected := r.URL.Query().Get("selected")
	selectedName := r.URL.Query().Get("selected_name")

	mfrs, err := app.services.manufacturers.List(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(fmt.Errorf("getEquipmentManufacturersFragment: %w", err))
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = manufacturersData{OrgID: orgID, Manufacturers: mfrs, SelectedID: selected, SelectedName: selectedName}
	return app.html.RenderFragment(w, r, http.StatusOK, fragments.EquipmentManufacturers, tmplData)
}
