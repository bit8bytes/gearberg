package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type equipmentCategoriesData struct {
	OrgID         string
	Categories    []categories.EquipmentCategory
	MaxCategories int
}

type equipmentCategoryData struct {
	OrgID    string
	Category *categories.EquipmentCategory
	ID       string
	ReturnTo string
}

func safeReturnTo(s string) string {
	if strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "//") {
		return s
	}
	return ""
}

func (app *application) getEquipmentCategories(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	cats, err := app.services.equipmentcategories.GetByOrgID(ctx, id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve equipment categories.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoriesData{
		OrgID:         id,
		Categories:    cats,
		MaxCategories: app.services.equipmentcategories.MaxCategories(),
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentCategoriesIndex, data)
}

func (app *application) getEquipmentCategoryNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")
	returnTo := safeReturnTo(r.URL.Query().Get("return_to"))
	data := app.html.TemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoryData{OrgID: id, ReturnTo: returnTo}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentCategoriesNew, data)
}

func (app *application) postEquipmentCategoryNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")
	returnTo := safeReturnTo(r.FormValue("return_to"))

	form, err := categories.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *categories.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentCategoryData{OrgID: id, ReturnTo: returnTo}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.equipmentcategories.Create(ctx, categories.CreateEquipmentCategory{
		ID:    ksuid.New().String(),
		OrgID: id,
		Name:  form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A category with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.equipmentcategories.MaxCategories()
			form.AddError("name", fmt.Sprintf("Category limit reached. Only %d categories allowed per orgs.", limit))
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to create equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	dest := "/orgs/" + url.PathEscape(id) + "/settings/equipment-categories"
	if returnTo != "" {
		dest = returnTo
	}
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is either a hard-coded path or validated by safeReturnTo (must start with "/" and not "//").
	return nil
}

func (app *application) getEquipmentCategory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	catID := r.PathValue("id")
	returnTo := safeReturnTo(r.URL.Query().Get("return_to"))

	category, err := app.services.equipmentcategories.GetByID(ctx, catID)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoryData{OrgID: orgID, Category: category, ID: catID, ReturnTo: returnTo}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentCategoriesDetail, data)
}

func (app *application) postEquipmentCategory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	catID := r.PathValue("id")
	returnTo := safeReturnTo(r.FormValue("return_to"))

	form, err := categories.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *categories.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentCategoryData{OrgID: orgID, ID: catID, ReturnTo: returnTo}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesDetail, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.equipmentcategories.Update(ctx, categories.UpdateEquipmentCategory{
		ID:   catID,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A category with this name already exists.")
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to update equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	dest := "/orgs/" + url.PathEscape(orgID) + "/settings/equipment-categories/" + url.PathEscape(catID)
	if returnTo != "" {
		dest = returnTo
	}
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is either a hard-coded path or validated by safeReturnTo (must start with "/" and not "//").
	return nil
}

func (app *application) postDeleteEquipmentCategory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	catID := r.PathValue("id")
	returnTo := safeReturnTo(r.FormValue("return_to"))

	if err := app.services.equipmentcategories.Delete(ctx, catID); err != nil {
		if errors.Is(err, database.ErrForeignKeyViolation) {
			category, fetchErr := app.services.equipmentcategories.GetByID(ctx, catID)
			if fetchErr != nil {
				return &httperr.Error{Error: fetchErr, Message: "Failed to retrieve equipment category.", Code: http.StatusInternalServerError}
			}
			f := &categories.Form{}
			f.AddError("delete", "Cannot delete: this category is assigned to one or more inventory items.")
			data := app.html.TemplateData(r)
			data.Form = f
			data.Data = equipmentCategoryData{OrgID: orgID, Category: category, ID: catID, ReturnTo: returnTo}
			return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesDetail, data)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	dest := "/orgs/" + url.PathEscape(orgID) + "/settings/equipment-categories"
	if returnTo != "" {
		dest = returnTo
	}
	http.Redirect(w, r, dest, http.StatusSeeOther) //nolint:gosec // dest is either a hard-coded path or validated by safeReturnTo (must start with "/" and not "//").
	return nil
}
