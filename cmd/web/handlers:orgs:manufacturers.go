package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs/manufacturers"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/segmentio/ksuid"
)

type manufacturersData struct {
	OrgID            string
	Manufacturers    []manufacturers.Manufacturer
	MaxManufacturers int
}

type manufacturerData struct {
	OrgID        string
	Manufacturer *manufacturers.Manufacturer
	ID           string
}

func (app *application) getManufacturers(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	mfrs, err := app.services.manufacturers.GetByOrgID(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve manufacturers.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &manufacturers.Form{}
	data.Data = manufacturersData{
		OrgID:            id,
		Manufacturers:    mfrs,
		MaxManufacturers: app.services.manufacturers.MaxManufacturers(),
	}
	return app.html.Render(w, r, http.StatusOK, pages.ManufacturersIndex, data)
}

func (app *application) getManufacturerNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")
	data := app.html.TemplateData(r)
	data.Form = &manufacturers.Form{}
	data.Data = manufacturerData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.ManufacturersNew, data)
}

func (app *application) postManufacturerNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := manufacturers.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *manufacturers.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = manufacturerData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ManufacturersNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.manufacturers.Create(ctx, manufacturers.CreateManufacturer{
		ID:    ksuid.New().String(),
		OrgID: id,
		Name:  form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A manufacturer with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.manufacturers.MaxManufacturers()
			form.AddError("name", fmt.Sprintf("Manufacturer limit reached. Only %d manufacturers allowed per org.", limit))
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to create manufacturer.",
			Code:    http.StatusInternalServerError,
		}
	}

	dest := "/orgs/" + url.PathEscape(id) + "/settings/manufacturers"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec
	return nil
}

func (app *application) getManufacturer(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	mfrID := r.PathValue("id")

	manufacturer, err := app.services.manufacturers.GetByID(ctx, mfrID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Manufacturer not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve manufacturer.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &manufacturers.Form{}
	data.Data = manufacturerData{OrgID: orgID, Manufacturer: manufacturer, ID: mfrID}
	return app.html.Render(w, r, http.StatusOK, pages.ManufacturersDetail, data)
}

func (app *application) postManufacturer(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	mfrID := r.PathValue("id")

	form, err := manufacturers.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	mfr, err := app.services.manufacturers.GetByID(ctx, mfrID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Manufacturer not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve manufacturer.", Code: http.StatusInternalServerError}
	}

	reRender := func(f *manufacturers.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = manufacturerData{OrgID: mfr.OrgID, Manufacturer: mfr, ID: mfrID}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ManufacturersDetail, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.manufacturers.Update(ctx, manufacturers.UpdateManufacturer{
		ID:   mfrID,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A manufacturer with this name already exists.")
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to update manufacturer.",
			Code:    http.StatusInternalServerError,
		}
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
		if errors.Is(err, database.ErrForeignKeyViolation) {
			manufacturer, fetchErr := app.services.manufacturers.GetByID(ctx, mfrID)
			if fetchErr != nil {
				if errors.Is(fetchErr, database.ErrNotFound) {
					return &httperr.Error{Error: fetchErr, Message: "Manufacturer not found.", Code: http.StatusNotFound}
				}
				return &httperr.Error{Error: fetchErr, Message: "Failed to retrieve manufacturer.", Code: http.StatusInternalServerError}
			}
			f := &manufacturers.Form{}
			f.AddError("delete", "Cannot delete: this manufacturer is assigned to one or more inventory items.")
			data := app.html.TemplateData(r)
			data.Form = f
			data.Data = manufacturerData{OrgID: orgID, Manufacturer: manufacturer, ID: mfrID}
			return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.ManufacturersDetail, data)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete manufacturer.",
			Code:    http.StatusInternalServerError,
		}
	}

	dest := "/orgs/" + url.PathEscape(orgID) + "/settings/manufacturers"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec
	return nil
}
