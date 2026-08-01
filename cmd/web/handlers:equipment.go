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
	"strings"
	"time"

	"github.com/bit8bytes/gearberg/internal/barcodes"
	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/categories"
	"github.com/bit8bytes/gearberg/internal/httperr"
	imgpkg "github.com/bit8bytes/gearberg/internal/image"
	"github.com/bit8bytes/gearberg/internal/money"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/internal/templates/fragments"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/internal/uid"
	"github.com/bit8bytes/gearberg/pkg/htmx"
)

type equipmentPartOfData struct {
	OrgID  string
	PartOf []equipment.PartOf
}

func (app *application) getEquipmentPartOfFragment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if !htmx.IsRequest(r) {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID), http.StatusSeeOther)
		return nil
	}

	partOf, err := app.services.equipment.ListContainers(ctx, itemID)
	if err != nil {
		return &httperr.Error{
			Error:   fmt.Errorf("getEquipmentPartOfFragment: %w", err),
			Message: "Failed to load container memberships.",
			Code:    http.StatusInternalServerError,
		}
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = equipmentPartOfData{OrgID: orgID, PartOf: partOf}
	return app.html.RenderFragment(w, r, http.StatusOK, fragments.EquipmentPartOf, tmplData)
}

type equipmentData struct {
	OrgID       string
	Categories  []categories.EquipmentCategory
	Inventories []equipment.Equipment
	Filtered    bool
	Query       string
	Category    string
	Sort        string
	PageBaseURL template.URL
	PrintURL    template.URL
	Pagination  pagination.Metadata
}

func (app *application) getEquipment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id, err := uid.Parse(r.PathValue("org_id"))
	if err != nil {
		return &httperr.Error{Error: err, Message: "Invalid organization ID.", Code: http.StatusBadRequest}
	}

	cats, err := app.services.equipment.ListCategories(ctx, id)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")
	sort := qs.Get("sort")

	page, err := strconv.Atoi(qs.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	f := pagination.Filters{
		Page:     page,
		PageSize: 25,
	}

	items, meta, err := app.services.equipment.GetFiltered(ctx, id, query, category, f)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory.", Code: http.StatusInternalServerError}
	}

	app.resolveEquipmentURLs(items)

	data := app.html.TemplateData(r)
	data.Data = equipmentData{
		OrgID:       id,
		Categories:  cats,
		Inventories: items,
		Filtered:    query != "" || category != "",
		Query:       query,
		Category:    category,
		Sort:        sort,
		PageBaseURL: template.URL(equipmentPageURL(id, query, category, sort)),  // #nosec G203
		PrintURL:    template.URL(equipmentPrintURL(id, query, category, sort)), // #nosec G203
		Pagination:  meta,
	}
	return app.html.Render(w, r, http.StatusOK, pages.Equipment, data)
}

type equipmentItemData struct {
	OrgID     string
	Item      *equipment.Equipment
	ID        string
	ActiveTab string
}

func (app *application) getEquipmentNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")
	data := app.html.TemplateData(r)
	data.Form = &equipment.NewForm{}
	data.Data = equipmentItemData{OrgID: id}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentNew, data)
}

func (app *application) postEquipmentNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	form, err := equipment.ParseNew(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *equipment.NewForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentItemData{OrgID: id}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	base := equipment.Base{
		OrgID:            id,
		UsageTypeID:      equipment.ParseUsage(form.UsageTypeID).ID(),
		Name:             form.Name,
		CategoryID:       form.CategoryID,
		CategoryName:     form.CategoryName,
		ManufacturerID:   form.ManufacturerID,
		ManufacturerName: form.ManufacturerName,
		LocationID:       form.LocationID,
		LocationName:     form.LocationName,
		HasContent:       form.HasContent,
		Notes:            form.Notes,
		Pricing:          form.ToPricing(),
	}

	eq, err := app.services.equipment.Create(ctx, equipment.CreateEquipment{
		TrackingType: equipment.TypeFromString(form.TypeID),
		Base:         base,
		TotalStock:   form.Count,
		UnitCount:    form.Count,
	})
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to create inventory item.", Code: http.StatusInternalServerError}
	}

	if appErr := app.processEquipmentImage(r, id, eq.ID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	if eq.TrackingType == equipment.Bulk {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment/"+url.PathEscape(eq.ID)+"/units", http.StatusSeeOther)
	}
	return nil
}

func (app *application) getEquipmentItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	data := app.html.TemplateData(r)
	f := equipment.DetailsFormFromEquipment(item)
	data.Form = &f
	data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, ActiveTab: "details"}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentDetail, data)
}

func (app *application) postEquipmentItemDetails(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseDetails(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if appErr := app.processEquipmentImage(r, orgID, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	if !form.Validate() {
		item, err := app.services.equipment.GetByID(ctx, itemID)
		if err != nil {
			return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
		}
		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, ActiveTab: "details"}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentDetail, data)
	}

	itemType := equipment.TypeFromString(form.TypeID)
	if err := app.services.equipment.UpdateDetails(ctx, equipment.UpdateEquipmentDetails{
		ID:               itemID,
		OrgID:            orgID,
		Type:             itemType,
		Name:             form.Name,
		CategoryID:       form.CategoryID,
		CategoryName:     form.CategoryName,
		ManufacturerID:   form.ManufacturerID,
		ManufacturerName: form.ManufacturerName,
		LocationID:       form.LocationID,
		LocationName:     form.LocationName,
		Notes:            form.Notes,
		TotalStock:       form.TotalStock,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"#save", http.StatusSeeOther)
	return nil
}

func (app *application) postEquipmentItemProperties(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseProperties(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		item, err := app.services.equipment.GetByID(ctx, itemID)
		if err != nil {
			return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
		}

		app.resolveItemURLs(item)

		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, ActiveTab: "properties"}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentProperties, data)
	}

	if err := app.services.equipment.UpdateProperties(ctx, equipment.UpdateEquipmentProperties{
		ID:         itemID,
		Properties: form.ToProperties(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/properties#save", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentItemPricing(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	data := app.html.TemplateData(r)
	f := equipment.PricingFormFromEquipment(item)
	data.Form = &f
	data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, ActiveTab: "pricing"}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentPricing, data)
}

func (app *application) postEquipmentItemPricing(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParsePricing(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		item, err := app.services.equipment.GetByID(ctx, itemID)
		if err != nil {
			return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
		}

		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, ActiveTab: "pricing"}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentPricing, data)
	}

	if err := app.services.equipment.UpdatePricing(ctx, equipment.UpdateEquipmentPricing{
		ID:      itemID,
		Pricing: form.ToPricing(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/pricing#save", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentItemProperties(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	data := app.html.TemplateData(r)
	f := equipment.PropertiesFormFromEquipment(item)
	data.Form = &f
	data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, ActiveTab: "properties"}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentProperties, data)
}

func (app *application) postEquipmentAddUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if _, err := app.services.equipment.AddUnit(ctx, equipment.AddUnit{
		OrgID:       orgID,
		EquipmentID: itemID,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to add unit.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) postEquipmentUpdateUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	form, err := equipment.ParseUnit(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		return app.renderEquipmentUnits(w, r, orgID, itemID)
	}

	if err := app.services.equipment.UpdateUnit(ctx, equipment.UpdateUnit{
		ID:                       unitID,
		SerialNumber:             form.SerialNumber,
		IsActive:                 form.StatusID,
		ManufacturerSerialNumber: form.ManufacturerSerialNumber,
		Remark:                   form.Remark,
		PurchasePrice:            form.ParsedPurchasePrice(),
		PurchasedAt:              form.PurchasedAt,
		NextInspectionAt:         form.NextInspectionAt,
	}); err != nil {
		if errors.Is(err, equipment.ErrConflict) {
			return app.renderEquipmentUnits(w, r, orgID, itemID)
		}
		return &httperr.Error{Error: err, Message: "Failed to update unit.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) postEquipmentBulkUpdateInspection(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	unitIDs := r.PostForm["unit_ids"]
	if len(unitIDs) == 0 {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
		return nil
	}

	var nextInspectionAt *int64
	if s := strings.TrimSpace(r.PostForm.Get("next_inspection_at")); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			v := t.UTC().Unix()
			nextInspectionAt = &v
		}
	}

	if err := app.services.equipment.BulkUpdateNextInspection(ctx, unitIDs, nextInspectionAt); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update units.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) postDeleteEquipmentUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	if err := app.services.equipment.DeleteUnit(ctx, unitID); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to delete unit.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

type equipmentUnitsData struct {
	OrgID     string
	ID        string
	ItemName  string
	ItemCode  int64
	Units     []equipment.Unit
	Item      *equipment.Equipment
	ActiveTab string
}

func (app *application) getEquipmentUnits(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	units, err := app.services.equipment.ListUnits(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve units.", Code: http.StatusInternalServerError}
	}

	data := app.html.TemplateData(r)
	data.Data = equipmentUnitsData{
		OrgID:     orgID,
		ID:        itemID,
		ItemName:  item.Name,
		Units:     units,
		Item:      item,
		ActiveTab: "units",
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentUnits, data)
}

func (app *application) getEquipmentUnitQR(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	unit, err := app.services.equipment.GetUnit(ctx, unitID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Unit not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve unit.", Code: http.StatusInternalServerError}
	}

	png, err := barcodes.QR(unit.SerialNumber)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to generate QR code.", Code: http.StatusInternalServerError}
	}

	filename := unitQRFilename(item.Name, unit.SerialNumber)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(png); err != nil { //nolint:gosec // png bytes are generated internally, not from user input
		return &httperr.Error{Error: err, Message: "Failed to write response.", Code: http.StatusInternalServerError}
	}
	return nil
}

func (app *application) getEquipmentUnitBarcode(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	unit, err := app.services.equipment.GetUnit(ctx, unitID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Unit not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve unit.", Code: http.StatusInternalServerError}
	}

	png, err := barcodes.Code128(unit.SerialNumber)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to generate barcode.", Code: http.StatusInternalServerError}
	}

	filename := unitQRFilename(item.Name, unit.SerialNumber)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(png); err != nil { //nolint:gosec // png bytes are generated internally, not from user input
		return &httperr.Error{Error: err, Message: "Failed to write response.", Code: http.StatusInternalServerError}
	}
	return nil
}

// unitQRFilename returns a safe PNG filename for a unit QR download.
// Format: {slugified-name}-{internal_id}.png.
func unitQRFilename(name, internalID string) string {
	slug := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, strings.ToLower(name))
	return slug + "-" + internalID + ".png"
}

// renderEquipmentUnits redirects to the units page, used when unit validation fails.
func (app *application) renderEquipmentUnits(w http.ResponseWriter, r *http.Request, orgID, itemID string) *httperr.Error {
	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

// processEquipmentImage stores and links an uploaded image. No-op when file is nil.
func (app *application) processEquipmentImage(r *http.Request, orgID, itemID string, file multipart.File, header *multipart.FileHeader) *httperr.Error {
	if file == nil {
		return nil
	}
	ctx := r.Context()

	result, err := imgpkg.Process(file)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Invalid image file.", Code: http.StatusUnprocessableEntity}
	}

	key := fmt.Sprintf("orgs/%s/equipment/%s", orgID, itemID)
	record, err := app.services.storageManager.Put(ctx, orgID, key, header.Filename, bytes.NewReader(result.Data), storage.Options{
		Size:        int64(len(result.Data)),
		ContentType: result.ContentType,
	})
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to store image.", Code: http.StatusInternalServerError}
	}

	if err := app.services.equipment.SetImage(ctx, equipment.SetImage{
		ID:              itemID,
		StorageObjectID: &record.ID,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to link image.", Code: http.StatusInternalServerError}
	}
	return nil
}

// equipmentPageURL builds the paginated base URL for the inventory list.
func equipmentPageURL(orgID, query, category, sort string) string {
	base := "/orgs/" + url.PathEscape(orgID) + "/equipment?"
	if category != "" {
		base += "category=" + url.QueryEscape(category) + "&"
	}
	if query != "" {
		base += "q=" + url.QueryEscape(query) + "&"
	}
	if sort != "" {
		base += "sort=" + url.QueryEscape(sort) + "&"
	}
	return base
}

// equipmentPrintURL builds the print URL with optional filter params.
func equipmentPrintURL(orgID, query, category, sort string) string {
	base := "/orgs/" + url.PathEscape(orgID) + "/equipment/print"
	sep := "?"
	if category != "" {
		base += sep + "category=" + url.QueryEscape(category)
		sep = "&"
	}
	if query != "" {
		base += sep + "q=" + url.QueryEscape(query)
		sep = "&"
	}
	if sort != "" {
		base += sep + "sort=" + url.QueryEscape(sort)
	}
	return base
}

// resolveItemURLs populates the ImageURL field from the storage object ID.
func (app *application) resolveItemURLs(item *equipment.Equipment) {
	if item.StorageObjectID != nil {
		item.ImageURL = app.services.storageManager.URL(*item.StorageObjectID)
	}
}

// resolveEquipmentURLs calls resolveItemURLs for each item in the slice.
func (app *application) resolveEquipmentURLs(items []equipment.Equipment) {
	for i := range items {
		app.resolveItemURLs(&items[i])
	}
}

func (app *application) postDeleteEquipmentItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if err := app.services.equipment.Delete(ctx, itemID); err != nil {
		if errors.Is(err, equipment.ErrInUse) {
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

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment", http.StatusSeeOther)
	return nil
}

type equipmentContentData struct {
	OrgID        string
	ID           string
	Item         *equipment.Equipment
	Content      []equipment.ContentItem
	AllEquipment []equipment.Equipment
	ActiveTab    string
}

func (app *application) getEquipmentContent(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetContentContainer(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		if errors.Is(err, equipment.ErrNoContentTab) {
			return &httperr.Error{Error: err, Message: "Content tab is not enabled for this item.", Code: http.StatusForbidden}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}
	app.resolveItemURLs(item)

	content, err := app.services.equipment.ListContent(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve content.", Code: http.StatusInternalServerError}
	}
	all, err := app.services.equipment.ListAll(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve equipment list.", Code: http.StatusInternalServerError}
	}

	data := app.html.TemplateData(r)
	data.Form = &equipment.ContentForm{}
	data.Data = equipmentContentData{OrgID: orgID, ID: itemID, Item: item, Content: content, AllEquipment: all, ActiveTab: "content"}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentContent, data)
}

func (app *application) renderEquipmentContentForm(w http.ResponseWriter, r *http.Request, orgID, itemID string, f *equipment.ContentForm, extraErr string) *httperr.Error {
	ctx := r.Context()
	item, err := app.services.equipment.GetContentContainer(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		if errors.Is(err, equipment.ErrNoContentTab) {
			return &httperr.Error{Error: err, Message: "Content tab is not enabled for this item.", Code: http.StatusForbidden}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}
	app.resolveItemURLs(item)
	content, err := app.services.equipment.ListContent(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve content.", Code: http.StatusInternalServerError}
	}
	all, err := app.services.equipment.ListAll(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve equipment list.", Code: http.StatusInternalServerError}
	}
	if extraErr != "" {
		f.AddError("assign", extraErr)
	}
	data := app.html.TemplateData(r)
	data.Form = f
	data.Data = equipmentContentData{OrgID: orgID, ID: itemID, Item: item, Content: content, AllEquipment: all, ActiveTab: "content"}
	return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentContent, data)
}

func (app *application) postEquipmentAssignContent(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseContent(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		return app.renderEquipmentContentForm(w, r, orgID, itemID, &form, "")
	}

	_, err = app.services.equipment.AssignContentByName(ctx, orgID, form.MemberName, equipment.AssignContent{
		EquipmentID: itemID,
		Quantity:    form.Quantity,
	})
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return app.renderEquipmentContentForm(w, r, orgID, itemID, &form, "No equipment found with that name.")
		}
		if errors.Is(err, equipment.ErrInvalidContent) {
			return app.renderEquipmentContentForm(w, r, orgID, itemID, &form, err.Error())
		}
		if errors.Is(err, equipment.ErrConflict) {
			return app.renderEquipmentContentForm(w, r, orgID, itemID, &form, "This item is already assigned as content.")
		}
		return app.renderEquipmentContentForm(w, r, orgID, itemID, &form, "Failed to assign content item.")
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/content", http.StatusSeeOther)
	return nil
}

func (app *application) postEquipmentRemoveContent(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")
	contentID := r.PathValue("content_id")

	if err := app.services.equipment.RemoveContent(ctx, contentID); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to remove content item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/content", http.StatusSeeOther)
	return nil
}

type equipmentPrintData struct {
	OrgID          string
	OrgDisplayName string
	Inventories    []equipment.Equipment
	Query          string
	Category       string
	PrintDate      string
	TotalCount     int
	Currency       money.Currency
	VatRate        money.VatRate
}

func (app *application) getEquipmentPrint(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	org, err := app.services.orgs.Get(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve organization.", Code: http.StatusInternalServerError}
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")
	showArchived := qs.Get("archived") == "true"

	filtered, err := app.services.equipment.ListAllFiltered(ctx, orgID, query, category, showArchived)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory.", Code: http.StatusInternalServerError}
	}

	app.resolveEquipmentURLs(filtered)

	orgSettings, err := app.services.orgsettings.Get(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve organization settings.", Code: http.StatusInternalServerError}
	}

	var currency money.Currency
	var vatRate money.VatRate
	if orgSettings != nil {
		currency = orgSettings.Currency
		vatRate = orgSettings.VatRate
	}

	data := app.html.TemplateData(r)
	data.Data = equipmentPrintData{
		OrgID:          orgID,
		OrgDisplayName: org.DisplayName,
		Inventories:    filtered,
		Query:          query,
		Category:       category,
		PrintDate:      time.Now().UTC().Format("2006-01-02"),
		TotalCount:     len(filtered),
		Currency:       currency,
		VatRate:        vatRate,
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentPrint, data)
}

