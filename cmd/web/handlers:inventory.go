package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/internal/companies/categories"
	"github.com/bit8bytes/gearberg/internal/inventory"
	"github.com/bit8bytes/gearberg/templates/pages"
)

type inventoryData struct {
	CompanyID   string
	Categories  []categories.EquipmentCategory
	Inventories []inventory.Inventory
}

func (app *application) getInventory(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	cats, err := app.services.inventory.ListCategories(ctx, id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve categories.",
			Code:    http.StatusInternalServerError,
		}
	}

	items, err := app.services.inventory.GetAll(ctx, id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve inventory.",
			Code:    http.StatusInternalServerError,
		}
	}

	data := app.newTemplateData(r)
	data.Data = inventoryData{CompanyID: id, Categories: cats, Inventories: items}
	return app.render(w, r, http.StatusOK, pages.Inventory, data)
}
