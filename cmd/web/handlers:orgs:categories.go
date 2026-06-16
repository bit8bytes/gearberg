package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
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
	data := app.html.TemplateData(r)
	data.Form = &categories.Form{Color: categories.DefaultColor}
	data.Data = equipmentCategoryData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentCategoriesNew, data)
}

func (app *application) postEquipmentCategoryNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := categories.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *categories.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentCategoryData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.equipmentcategories.Create(ctx, categories.CreateEquipmentCategory{
		ID:    ksuid.New().String(),
		OrgID: id,
		Name:  form.Name,
		Color: form.Color,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A category with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.equipmentcategories.MaxCategories()
			form.AddError("name", fmt.Sprintf("Category limit reached. Only %d categories allowed per org.", limit))
			return reRender(&form)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to create equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/settings/equipment-categories", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentCategory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	catID := r.PathValue("id")

	category, err := app.services.equipmentcategories.GetByID(ctx, catID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment category not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to retrieve equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.html.TemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoryData{OrgID: orgID, Category: category, ID: catID}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentCategoriesDetail, data)
}

func (app *application) postEquipmentCategory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	catID := r.PathValue("id")

	form, err := categories.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	category, err := app.services.equipmentcategories.GetByID(ctx, catID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment category not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve equipment category.", Code: http.StatusInternalServerError}
	}

	reRender := func(f *categories.Form) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentCategoryData{OrgID: orgID, Category: category, ID: catID}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesDetail, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.equipmentcategories.Update(ctx, categories.UpdateEquipmentCategory{
		ID:    catID,
		Name:  form.Name,
		Color: form.Color,
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

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/settings/equipment-categories/"+url.PathEscape(catID), http.StatusSeeOther)
	return nil
}

func (app *application) postDeleteEquipmentCategory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	catID := r.PathValue("id")

	if err := app.services.equipmentcategories.Delete(ctx, catID); err != nil {
		if errors.Is(err, database.ErrForeignKeyViolation) {
			category, fetchErr := app.services.equipmentcategories.GetByID(ctx, catID)
			if fetchErr != nil {
				if errors.Is(fetchErr, database.ErrNotFound) {
					return &httperr.Error{Error: fetchErr, Message: "Equipment category not found.", Code: http.StatusNotFound}
				}
				return &httperr.Error{Error: fetchErr, Message: "Failed to retrieve equipment category.", Code: http.StatusInternalServerError}
			}
			f := &categories.Form{}
			f.AddError("delete", "Cannot delete: this category is assigned to one or more inventory items.")
			data := app.html.TemplateData(r)
			data.Form = f
			data.Data = equipmentCategoryData{OrgID: orgID, Category: category, ID: catID}
			return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesDetail, data)
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/settings/equipment-categories", http.StatusSeeOther)
	return nil
}
