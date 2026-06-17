package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/locations"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/segmentio/ksuid"
)

type locationsData struct {
	OrgID        string
	Locations    []locations.Location
	MaxLocations int
}

type locationData struct {
	OrgID    string
	Location *locations.Location
	ID       string
}

func (app *application) getLocations(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	locs, err := app.services.locations.GetByOrgID(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve locations.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &locations.Form{}
	data.Data = locationsData{
		OrgID:        id,
		Locations:    locs,
		MaxLocations: app.services.locations.MaxLocations(),
	}
	return app.html.Render(w, r, http.StatusOK, pages.LocationsIndex, data)
}

func (app *application) getLocationNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")
	data := app.html.TemplateData(r)
	data.Form = &locations.Form{}
	data.Data = locationData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.LocationsNew, data)
}

func (app *application) postLocationNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := locations.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *locations.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = locationData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.LocationsNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.locations.Create(ctx, locations.CreateLocation{
		ID:    ksuid.New().String(),
		OrgID: id,
		Name:  form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A location with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.locations.MaxLocations()
			form.AddError("name", fmt.Sprintf("Location limit reached. Only %d locations allowed per org.", limit))
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to create location.",
			Code:    http.StatusInternalServerError,
		}
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
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Location not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve location.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &locations.Form{}
	data.Data = locationData{OrgID: orgID, Location: loc, ID: locID}
	return app.html.Render(w, r, http.StatusOK, pages.LocationsDetail, data)
}

func (app *application) postLocation(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	locID := r.PathValue("id")

	form, err := locations.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	loc, err := app.services.locations.GetByID(ctx, locID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Location not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve location.", Code: http.StatusInternalServerError}
	}

	reRender := func(f *locations.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = locationData{OrgID: loc.OrgID, Location: loc, ID: locID}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.LocationsDetail, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.locations.Update(ctx, locations.UpdateLocation{
		ID:   locID,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A location with this name already exists.")
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to update location.",
			Code:    http.StatusInternalServerError,
		}
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
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete location.",
			Code:    http.StatusInternalServerError,
		}
	}

	dest := "/orgs/" + url.PathEscape(orgID) + "/settings/locations"
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is a hard-coded path with org_id from path value.
	return nil
}
