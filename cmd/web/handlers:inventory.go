package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/internal/companies/categories"
	imgpkg "github.com/bit8bytes/gearberg/internal/image"
	"github.com/bit8bytes/gearberg/internal/inventory"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type inventoryData struct {
	CompanyID   string
	Categories  []categories.EquipmentCategory
	Inventories []inventory.Inventory
	Filtered    bool
	SortBy      string
	SortDir     string
	SortBaseURL template.URL
}

func parseSortParams(q url.Values) (sortBy, sortDir string) {
	sortBy = q.Get("sort")
	switch sortBy {
	case "name", "category", "stock":
	default:
		return "", ""
	}
	if q.Get("dir") == "desc" {
		return sortBy, "desc"
	}
	return sortBy, "asc"
}

type inventoryItemData struct {
	CompanyID  string
	Item       *inventory.Inventory
	ID         string
	Categories []categories.EquipmentCategory
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

	selectedCategories := r.URL.Query()["category"]

	sortBy, sortDir := parseSortParams(r.URL.Query())

	items, err := app.services.inventory.GetFiltered(ctx, id, selectedCategories, sortBy, sortDir)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to retrieve inventory.",
			Code:    http.StatusInternalServerError,
		}
	}

	for i := range items {
		if items[i].StorageObjectID != nil {
			items[i].ImageURL = app.services.storageManager.URL(*items[i].StorageObjectID)
		}
	}

	var baseURL strings.Builder
	baseURL.WriteString("/companies/" + url.PathEscape(id) + "/inventory?")
	for i, c := range selectedCategories {
		if i > 0 {
			baseURL.WriteString("&")
		}
		baseURL.WriteString("category=" + url.QueryEscape(c))
	}
	if len(selectedCategories) > 0 {
		baseURL.WriteString("&")
	}

	data := app.newTemplateData(r)
	data.Data = inventoryData{
		CompanyID:   id,
		Categories:  cats,
		Inventories: items,
		Filtered:    len(selectedCategories) > 0,
		SortBy:      sortBy,
		SortDir:     sortDir,
		SortBaseURL: template.URL(baseURL.String()), // #nosec G203 — baseURL is built with url.PathEscape/url.QueryEscape
	}
	return app.render(w, r, http.StatusOK, pages.Inventory, data)
}

func (app *application) getInventoryNew(w http.ResponseWriter, r *http.Request) *appError {
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

	data := app.newTemplateData(r)
	data.Form = &inventory.Form{}
	data.Data = inventoryItemData{CompanyID: id, Categories: cats}
	return app.render(w, r, http.StatusOK, pages.InventoryNew, data)
}

func (app *application) postInventoryNew(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := r.PathValue("company_id")

	form, err := inventory.Parse(r)
	if err != nil {
		return &appError{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *inventory.Form) *appError {
		cats, catErr := app.services.inventory.ListCategories(ctx, id)
		if catErr != nil {
			return &appError{Error: catErr, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
		}
		data := app.newTemplateData(r)
		data.Form = f
		data.Data = inventoryItemData{CompanyID: id, Categories: cats}
		return app.render(w, r, http.StatusUnprocessableEntity, pages.InventoryNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	itemID := ksuid.New().String()
	_, err = app.services.inventory.Create(ctx, inventory.CreateInventory{
		ID:         itemID,
		CompanyID:  id,
		Name:       form.Name,
		CategoryID: form.CategoryID,
		TotalStock: form.TotalStockInt64(),
		Notes:      form.Notes,
	})
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to create inventory item.",
			Code:    http.StatusInternalServerError,
		}
	}

	if form.Image != nil {
		record, appErr := app.storeInventoryImage(r, id, itemID, &form)
		if appErr != nil {
			return appErr
		}
		if appErr := app.linkInventoryImage(r, itemID, record.ID); appErr != nil {
			return appErr
		}
	}

	http.Redirect(w, r, "/companies/"+url.PathEscape(id)+"/inventory", http.StatusSeeOther)
	return nil
}

func (app *application) getInventoryItem(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	companyID := r.PathValue("company_id")
	itemID := r.PathValue("id")

	item, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &appError{Error: err, Message: "Inventory item not found.", Code: http.StatusNotFound}
		}
		return &appError{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}

	cats, err := app.services.inventory.ListCategories(ctx, companyID)
	if err != nil {
		return &appError{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	data := app.newTemplateData(r)
	data.Form = &inventory.Form{}
	data.Data = inventoryItemData{CompanyID: companyID, Item: item, ID: itemID, Categories: cats}
	return app.render(w, r, http.StatusOK, pages.InventoryDetail, data)
}

func (app *application) postInventoryItem(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	companyID := r.PathValue("company_id")
	itemID := r.PathValue("id")

	form, err := inventory.Parse(r)
	if err != nil {
		return &appError{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if form.Image != nil {
		record, appErr := app.storeInventoryImage(r, companyID, itemID, &form)
		if appErr != nil {
			return appErr
		}
		if appErr := app.linkInventoryImage(r, itemID, record.ID); appErr != nil {
			return appErr
		}
	}

	if !form.Validate() {
		return app.renderInventoryEdit(w, r, companyID, itemID, &form)
	}

	item, err := app.services.inventory.Update(ctx, inventory.UpdateInventory{
		ID:         itemID,
		Name:       form.Name,
		CategoryID: form.CategoryID,
		TotalStock: form.TotalStockInt64(),
		Notes:      form.Notes,
	})
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to update inventory item.",
			Code:    http.StatusInternalServerError,
		}
	}

	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}

	cats, err := app.services.inventory.ListCategories(ctx, companyID)
	if err != nil {
		return &appError{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	data := app.newTemplateData(r)
	data.Form = &inventory.Form{}
	data.Data = inventoryItemData{CompanyID: companyID, Item: item, ID: itemID, Categories: cats}
	return app.render(w, r, http.StatusOK, pages.InventoryDetail, data)
}

// storeInventoryImage processes the uploaded image and writes it to storage.
// It does not link the image to the inventory item; call linkInventoryImage for that.
func (app *application) storeInventoryImage(r *http.Request, companyID, itemID string, form *inventory.Form) (*storage.Object, *appError) {
	ctx := r.Context()

	result, err := imgpkg.Process(form.Image)
	if err != nil {
		return nil, &appError{Error: err, Message: "Invalid image file.", Code: http.StatusUnprocessableEntity}
	}

	key := fmt.Sprintf("company/%s/inventory/%s", companyID, itemID)
	record, err := app.services.storageManager.Put(ctx, companyID, key, form.ImageHeader.Filename, bytes.NewReader(result.Data), storage.Options{
		Size:        int64(len(result.Data)),
		ContentType: result.ContentType,
	})
	if err != nil {
		return nil, &appError{Error: err, Message: "Failed to store image.", Code: http.StatusInternalServerError}
	}

	return record, nil
}

// renderInventoryEdit re-fetches the item and categories and renders the edit
// form with 422 Unprocessable Entity, used when validation fails.
func (app *application) renderInventoryEdit(w http.ResponseWriter, r *http.Request, companyID, itemID string, f *inventory.Form) *appError {
	ctx := r.Context()
	item, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		return &appError{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}
	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}
	cats, err := app.services.inventory.ListCategories(ctx, companyID)
	if err != nil {
		return &appError{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}
	data := app.newTemplateData(r)
	data.Form = f
	data.Data = inventoryItemData{CompanyID: companyID, Item: item, ID: itemID, Categories: cats}
	return app.render(w, r, http.StatusUnprocessableEntity, pages.InventoryDetail, data)
}

// linkInventoryImage associates a storage object with an inventory item.
func (app *application) linkInventoryImage(r *http.Request, itemID, storageObjectID string) *appError {
	if err := app.services.inventory.SetImage(r.Context(), inventory.SetImage{
		ID:              itemID,
		StorageObjectID: &storageObjectID,
	}); err != nil {
		return &appError{Error: err, Message: "Failed to link image.", Code: http.StatusInternalServerError}
	}
	return nil
}

func (app *application) postDeleteInventoryItem(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	companyID := r.PathValue("company_id")
	itemID := r.PathValue("id")

	if err := app.services.inventory.Delete(ctx, itemID); err != nil {
		if errors.Is(err, database.ErrForeignKeyViolation) {
			return &appError{
				Error:   err,
				Message: "Cannot delete an inventory item that is part of an active rental.",
				Code:    http.StatusConflict,
			}
		}
		return &appError{
			Error:   err,
			Message: "Failed to delete inventory item.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/companies/"+url.PathEscape(companyID)+"/inventory", http.StatusSeeOther)
	return nil
}
