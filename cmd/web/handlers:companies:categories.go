package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/internal/companies/categories"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type equipmentCategoriesData struct {
	CompanyID     string
	Categories    []categories.EquipmentCategory
	MaxCategories int
}

type equipmentCategoryData struct {
	CompanyID string
	Category  *categories.EquipmentCategory
	ID        string
}

func (app *application) getEquipmentCategories(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	cats, err := app.services.equipmentcategories.GetByCompanyID(ctx, id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve equipment categories.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoriesData{
		CompanyID:     id,
		Categories:    cats,
		MaxCategories: app.services.equipmentcategories.MaxCategories(),
	}
	return app.render(w, r, http.StatusOK, pages.EquipmentCategoriesIndex, data)
}

func (app *application) getEquipmentCategoryNew(w http.ResponseWriter, r *http.Request) *appError {
	id := r.PathValue("company_id")
	data := app.newTemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoryData{CompanyID: id}
	return app.render(w, r, http.StatusOK, pages.EquipmentCategoriesNew, data)
}

func (app *application) postEquipmentCategoryNew(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	form, err := categories.Parse(r)
	if err != nil {
		return &appError{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *categories.Form) *appError {
		data := app.newTemplateData(r)
		data.Form = f
		data.Data = equipmentCategoryData{CompanyID: id}
		return app.render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	_, err = app.services.equipmentcategories.Create(ctx, categories.CreateEquipmentCategory{
		ID:        ksuid.New().String(),
		CompanyID: id,
		Name:      form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A category with this name already exists.")
			return reRender(&form)
		}
		if errors.Is(err, database.ErrLimitExceeded) {
			limit := app.services.equipmentcategories.MaxCategories()
			form.AddError("name", fmt.Sprintf("Category limit reached. Only %d categories allowed per company.", limit))
			return reRender(&form)
		}
		return &appError{
			Error:   err,
			Message: "Failed to create equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/companies/"+url.PathEscape(id)+"/settings/equipment-categories", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentCategory(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	companyID := r.PathValue("company_id")
	catID := r.PathValue("id")

	category, err := app.services.equipmentcategories.GetByID(ctx, catID)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoryData{CompanyID: companyID, Category: category, ID: catID}
	return app.render(w, r, http.StatusOK, pages.EquipmentCategoriesDetail, data)
}

func (app *application) postEquipmentCategory(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	companyID := r.PathValue("company_id")
	catID := r.PathValue("id")

	form, err := categories.Parse(r)
	if err != nil {
		return &appError{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *categories.Form) *appError {
		data := app.newTemplateData(r)
		data.Form = f
		data.Data = equipmentCategoryData{CompanyID: companyID, ID: catID}
		return app.render(w, r, http.StatusUnprocessableEntity, pages.EquipmentCategoriesDetail, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	category, err := app.services.equipmentcategories.Update(ctx, categories.UpdateEquipmentCategory{
		ID:   catID,
		Name: form.Name,
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			form.AddError("name", "A category with this name already exists.")
			return reRender(&form)
		}
		return &appError{
			Error:   err,
			Message: "Failed to update equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Form = &categories.Form{}
	data.Data = equipmentCategoryData{CompanyID: companyID, Category: category, ID: catID}
	return app.render(w, r, http.StatusOK, pages.EquipmentCategoriesDetail, data)
}

func (app *application) postDeleteEquipmentCategory(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	companyID := r.PathValue("company_id")
	catID := r.PathValue("id")

	if err := app.services.equipmentcategories.Delete(ctx, catID); err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to delete equipment category.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/companies/"+url.PathEscape(companyID)+"/settings/equipment-categories", http.StatusSeeOther)
	return nil
}
