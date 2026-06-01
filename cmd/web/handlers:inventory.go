package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bit8bytes/gearberg/database"
	"github.com/bit8bytes/gearberg/internal/httperr"
	imgpkg "github.com/bit8bytes/gearberg/internal/image"
	"github.com/bit8bytes/gearberg/internal/inventory"
	invtypes "github.com/bit8bytes/gearberg/internal/inventory/types"
	"github.com/bit8bytes/gearberg/internal/orgs/categories"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/templates/pages"
	"github.com/segmentio/ksuid"
)

type inventoryData struct {
	OrgID       string
	Categories  []categories.EquipmentCategory
	Inventories []inventory.Inventory
	Filtered    bool
	Query       string
	Category    string
	PageBaseURL template.URL
	Pagination  pagination.Metadata
}

type inventoryItemData struct {
	OrgID      string
	Item       *inventory.Inventory
	ID         string
	Categories []categories.EquipmentCategory
	Currency   string
}

type inventoryUnitsData struct {
	OrgID    string
	ItemID   string
	ItemName string
	Units    []inventory.Unit
}

func (app *application) getInventory(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	cats, err := app.services.inventory.ListCategories(ctx, id)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")

	page, err := strconv.Atoi(qs.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	f := pagination.Filters{
		Page:     page,
		PageSize: 25,
	}

	items, meta, err := app.services.inventory.GetFiltered(ctx, id, query, category, f)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory.", Code: http.StatusInternalServerError}
	}

	for i := range items {
		if items[i].StorageObjectID != nil {
			items[i].ImageURL = app.services.storageManager.URL(*items[i].StorageObjectID)
		}
	}

	pageBaseURL := "/orgs/" + url.PathEscape(id) + "/inventory?"
	if category != "" {
		pageBaseURL += "category=" + url.QueryEscape(category) + "&"
	}
	if query != "" {
		pageBaseURL += "q=" + url.QueryEscape(query) + "&"
	}

	data := app.html.TemplateData(r)
	data.Data = inventoryData{
		OrgID:       id,
		Categories:  cats,
		Inventories: items,
		Filtered:    query != "" || category != "",
		Query:       query,
		Category:    category,
		PageBaseURL: template.URL(pageBaseURL), // #nosec G203
		Pagination:  meta,
	}
	return app.html.Render(w, r, http.StatusOK, pages.Inventory, data)
}

func (app *application) getInventoryNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	cats, err := app.services.inventory.ListCategories(ctx, id)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, id)
	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	data.Form = &inventory.NewForm{}
	data.Data = inventoryItemData{OrgID: id, Categories: cats, Currency: currency}
	return app.html.Render(w, r, http.StatusOK, pages.InventoryNew, data)
}

func (app *application) postInventoryNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := inventory.ParseNew(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *inventory.NewForm) *httperr.Error {
		cats, catErr := app.services.inventory.ListCategories(ctx, id)
		if catErr != nil {
			return &httperr.Error{Error: catErr, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
		}
		orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, id)
		currency := ""
		if orgSettings != nil {
			currency = orgSettings.Currency
		}
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = inventoryItemData{OrgID: id, Categories: cats, Currency: currency}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.InventoryNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	itemID := ksuid.New().String()
	base := inventory.Base{
		ID:            itemID,
		OrgID:         id,
		UsageTypeID:   invtypes.ParseUsage(form.UsageTypeID).ID(),
		Name:          form.Name,
		CategoryID:    form.CategoryID,
		Code:          form.Code,
		PurchasePrice: form.PurchasePriceCents(),
		RentalPrice:   form.RentalPriceCents(),
		Notes:         form.Notes,
	}

	if form.TypeID == "bulk" {
		return app.createBulkItem(w, r, id, itemID, base, &form)
	}
	return app.createSerializedItem(w, r, id, itemID, base, &form)
}

func (app *application) createBulkItem(w http.ResponseWriter, r *http.Request, orgID, itemID string, base inventory.Base, form *inventory.NewForm) *httperr.Error {
	ctx := r.Context()
	if _, err := app.services.inventory.CreateBulk(ctx, inventory.CreateBulkInventory{
		Base:       base,
		TotalStock: form.CountInt64(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to create inventory item.", Code: http.StatusInternalServerError}
	}
	if appErr := app.processInventoryImage(r, orgID, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}
	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/inventory", http.StatusSeeOther)
	return nil
}

func (app *application) createSerializedItem(w http.ResponseWriter, r *http.Request, orgID, itemID string, base inventory.Base, form *inventory.NewForm) *httperr.Error {
	ctx := r.Context()
	count := form.CountInt64()
	units := make([]inventory.CreateUnit, count)
	for i := range units {
		units[i] = inventory.CreateUnit{
			ID:          ksuid.New().String(),
			InventoryID: itemID,
			UnitNumber:  int64(i + 1),
		}
	}
	tx, txErr := app.db.BeginTx(ctx, nil)
	if txErr != nil {
		return &httperr.Error{Error: txErr, Message: "Failed to start transaction.", Code: http.StatusInternalServerError}
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := app.services.inventory.CreateSerialized(ctx, tx, inventory.CreateSerializedInventory{
		Base:  base,
		Units: units,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to create inventory item.", Code: http.StatusInternalServerError}
	}
	if err := tx.Commit(); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to save inventory item.", Code: http.StatusInternalServerError}
	}
	if appErr := app.processInventoryImage(r, orgID, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}
	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/inventory/"+url.PathEscape(itemID), http.StatusSeeOther)
	return nil
}

func (app *application) getInventoryItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Inventory item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}

	cats, err := app.services.inventory.ListCategories(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, orgID)
	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	data.Form = &inventory.Form{}
	data.Data = inventoryItemData{OrgID: orgID, Item: item, ID: itemID, Categories: cats, Currency: currency}
	page := pages.InventoryDetailBulk
	if item.TypeID == invtypes.Serialized.ID() {
		page = pages.InventoryDetailSerialized
	}
	return app.html.Render(w, r, http.StatusOK, page, data)
}

func (app *application) postInventoryItemBulk(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := inventory.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if appErr := app.processInventoryImage(r, orgID, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	if !form.ValidateBulk() {
		return app.renderInventoryEdit(w, r, orgID, itemID, &form)
	}

	item, err := app.services.inventory.Update(ctx, inventory.UpdateInventory{
		ID:            itemID,
		Name:          form.Name,
		CategoryID:    form.CategoryID,
		Code:          form.Code,
		TotalStock:    form.TotalStockInt64(),
		PurchasePrice: form.PurchasePriceCents(),
		RentalPrice:   form.RentalPriceCents(),
		Notes:         form.Notes,
	})
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}

	cats, err := app.services.inventory.ListCategories(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, orgID)
	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	data.Form = &inventory.Form{}
	data.Data = inventoryItemData{OrgID: orgID, Item: item, ID: itemID, Categories: cats, Currency: currency}
	return app.html.Render(w, r, http.StatusOK, pages.InventoryDetailBulk, data)
}

func (app *application) postInventoryItemSerialized(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := inventory.Parse(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if appErr := app.processInventoryImage(r, orgID, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	if !form.ValidateSerialized() {
		return app.renderInventoryEdit(w, r, orgID, itemID, &form)
	}

	// total_stock is derived from unit count; preserve the current value.
	current, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	item, err := app.services.inventory.Update(ctx, inventory.UpdateInventory{
		ID:            itemID,
		Name:          form.Name,
		CategoryID:    form.CategoryID,
		Code:          form.Code,
		TotalStock:    current.TotalStock,
		PurchasePrice: form.PurchasePriceCents(),
		RentalPrice:   form.RentalPriceCents(),
		Notes:         form.Notes,
	})
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}

	cats, err := app.services.inventory.ListCategories(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, orgID)
	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	data.Form = &inventory.Form{}
	data.Data = inventoryItemData{OrgID: orgID, Item: item, ID: itemID, Categories: cats, Currency: currency}
	return app.html.Render(w, r, http.StatusOK, pages.InventoryDetailSerialized, data)
}

func (app *application) postInventoryAddUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := inventory.ParseUnit(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		return app.renderInventoryUnits(w, r, orgID, itemID)
	}

	if _, err := app.services.inventory.AddUnit(ctx, inventory.AddUnit{
		ID:          ksuid.New().String(),
		InventoryID: itemID,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to add unit.", Code: http.StatusInternalServerError}
	}

	// After add, update the unit's fields if provided.
	// AddUnit creates an empty unit; we patch it immediately if serial/date were supplied.
	// This avoids a separate "edit after add" round-trip for users who fill the form.
	// We don't know the new unit ID here, so we list units and update the last one.
	if form.SerialNumber != "" || form.NextInspectionAt != "" {
		units, listErr := app.services.inventory.ListUnits(ctx, itemID)
		if listErr == nil && len(units) > 0 {
			last := units[len(units)-1]
			_ = app.services.inventory.UpdateUnit(ctx, inventory.UpdateUnit{
				ID:               last.ID,
				SerialNumber:     form.SerialNumber,
				NextInspectionAt: form.NextInspectionAtUnix(),
			})
		}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/inventory/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) postInventoryUpdateUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	form, err := inventory.ParseUnit(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		return app.renderInventoryUnits(w, r, orgID, itemID)
	}

	if err := app.services.inventory.UpdateUnit(ctx, inventory.UpdateUnit{
		ID:               unitID,
		SerialNumber:     form.SerialNumber,
		NextInspectionAt: form.NextInspectionAtUnix(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update unit.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/inventory/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) postDeleteInventoryUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	if err := app.services.inventory.DeleteUnit(ctx, unitID, itemID); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to delete unit.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/inventory/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) getInventoryUnits(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Inventory item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	units, err := app.services.inventory.ListUnits(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve units.", Code: http.StatusInternalServerError}
	}

	data := app.html.TemplateData(r)
	data.Data = inventoryUnitsData{OrgID: orgID, ItemID: itemID, ItemName: item.Name, Units: units}
	return app.html.Render(w, r, http.StatusOK, pages.InventoryUnits, data)
}

// renderInventoryUnits re-fetches units and renders the units page with 422, used when unit validation fails.
func (app *application) renderInventoryUnits(w http.ResponseWriter, r *http.Request, orgID, itemID string) *httperr.Error {
	ctx := r.Context()

	item, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	units, err := app.services.inventory.ListUnits(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve units.", Code: http.StatusInternalServerError}
	}

	data := app.html.TemplateData(r)
	data.Data = inventoryUnitsData{OrgID: orgID, ItemID: itemID, ItemName: item.Name, Units: units}
	return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.InventoryUnits, data)
}

// processInventoryImage stores and links an uploaded image. No-op when file is nil.
func (app *application) processInventoryImage(r *http.Request, orgID, itemID string, file multipart.File, header *multipart.FileHeader) *httperr.Error {
	if file == nil {
		return nil
	}
	record, appErr := app.storeInventoryImage(r, orgID, itemID, file, header)
	if appErr != nil {
		return appErr
	}
	return app.linkInventoryImage(r, itemID, record.ID)
}

// storeInventoryImage processes an uploaded image and writes it to storage.
// It does not link the image to the inventory item; call linkInventoryImage for that.
func (app *application) storeInventoryImage(r *http.Request, orgID, itemID string, file multipart.File, header *multipart.FileHeader) (*storage.Object, *httperr.Error) {
	ctx := r.Context()

	result, err := imgpkg.Process(file)
	if err != nil {
		return nil, &httperr.Error{Error: err, Message: "Invalid image file.", Code: http.StatusUnprocessableEntity}
	}

	key := fmt.Sprintf("orgs/%s/inventory/%s", orgID, itemID)
	record, err := app.services.storageManager.Put(ctx, orgID, key, header.Filename, bytes.NewReader(result.Data), storage.Options{
		Size:        int64(len(result.Data)),
		ContentType: result.ContentType,
	})
	if err != nil {
		return nil, &httperr.Error{Error: err, Message: "Failed to store image.", Code: http.StatusInternalServerError}
	}

	return record, nil
}

// renderInventoryEdit re-fetches the item and categories and renders the edit
// form with 422 Unprocessable Entity, used when validation fails.
// f may be nil for serialized items where the item form is valid but a unit form failed.
func (app *application) renderInventoryEdit(w http.ResponseWriter, r *http.Request, orgID, itemID string, f *inventory.Form) *httperr.Error {
	ctx := r.Context()
	item, err := app.services.inventory.GetByID(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}
	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}
	cats, err := app.services.inventory.ListCategories(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, orgID)
	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	if f == nil {
		f = &inventory.Form{}
	}
	data := app.html.TemplateData(r)
	data.Form = f
	data.Data = inventoryItemData{OrgID: orgID, Item: item, ID: itemID, Categories: cats, Currency: currency}
	page := pages.InventoryDetailBulk
	if item.TypeID == invtypes.Serialized.ID() {
		page = pages.InventoryDetailSerialized
	}
	return app.html.Render(w, r, http.StatusUnprocessableEntity, page, data)
}

// linkInventoryImage associates a storage object with an inventory item.
func (app *application) linkInventoryImage(r *http.Request, itemID, storageObjectID string) *httperr.Error {
	if err := app.services.inventory.SetImage(r.Context(), inventory.SetImage{
		ID:              itemID,
		StorageObjectID: &storageObjectID,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to link image.", Code: http.StatusInternalServerError}
	}
	return nil
}

func (app *application) postDeleteInventoryItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if err := app.services.inventory.Delete(ctx, itemID); err != nil {
		if errors.Is(err, database.ErrForeignKeyViolation) {
			return &httperr.Error{
				Error:   err,
				Message: "Cannot delete an inventory item that is part of an active rental.",
				Code:    http.StatusConflict,
			}
		}
		return &httperr.Error{
			Error:   err,
			Message: "Failed to delete inventory item.",
			Code:    http.StatusInternalServerError,
		}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/inventory", http.StatusSeeOther)
	return nil
}
