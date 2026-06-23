package main

import (
	"bytes"
	"context"
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
	"github.com/bit8bytes/gearberg/internal/database"
	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/categories"
	"github.com/bit8bytes/gearberg/internal/equipment/locations"
	"github.com/bit8bytes/gearberg/internal/equipment/manufacturers"
	"github.com/bit8bytes/gearberg/internal/httperr"
	imgpkg "github.com/bit8bytes/gearberg/internal/image"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/serial"
	"github.com/bit8bytes/gearberg/internal/storage"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/segmentio/ksuid"
)

type equipmentData struct {
	OrgID             string
	Categories        []categories.EquipmentCategory
	Inventories       []equipment.Equipment
	Filtered          bool
	Query             string
	Category          string
	Sort              string
	ShowArchived      bool
	PageBaseURL       template.URL
	PrintURL          template.URL
	ToggleArchivedURL template.URL
	Pagination        pagination.Metadata
}

type equipmentPrintData struct {
	OrgID          string
	OrgName        string
	Inventories    []equipment.Equipment
	Query          string
	Category       string
	PrintDate      string
	TotalCount     int
	Currency       string
	VatRateDisplay string
}

type equipmentItemData struct {
	OrgID         string
	Item          *equipment.Equipment
	ID            string
	Categories    []categories.EquipmentCategory
	Manufacturers []manufacturers.Manufacturer
	Locations     []locations.Location
	Currency      string
	PartOf        []equipment.PartOf
	ActiveTab     string
}

type equipmentUnitsData struct {
	OrgID      string
	ID         string
	ItemName   string
	ItemCode   int64
	Units      []equipment.Unit
	Item       *equipment.Equipment
	Categories []categories.EquipmentCategory
	PartOf     []equipment.PartOf
	ActiveTab  string
}

type equipmentContentData struct {
	OrgID        string
	ID           string
	Item         *equipment.Equipment
	Content      []equipment.ContentItem
	AllEquipment []equipment.Equipment
	PartOf       []equipment.PartOf
	ActiveTab    string
}

func (app *application) loadPartOf(ctx context.Context, equipmentID string) ([]equipment.PartOf, *httperr.Error) {
	partOf, err := app.services.equipment.ListContainers(ctx, equipmentID)
	if err != nil {
		return nil, &httperr.Error{Error: err, Message: "Failed to load container memberships.", Code: http.StatusInternalServerError}
	}
	return partOf, nil
}

func (app *application) getEquipment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	cats, err := app.services.equipment.ListCategories(ctx, id)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")
	sort := qs.Get("sort")
	showArchived := qs.Get("archived") == "true"

	page, err := strconv.Atoi(qs.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	f := pagination.Filters{
		Page:     page,
		PageSize: 25,
	}

	items, meta, err := app.services.equipment.GetFiltered(ctx, id, query, category, showArchived, f)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory.", Code: http.StatusInternalServerError}
	}

	app.resolveEquipmentURLs(items)

	data := app.html.TemplateData(r)
	data.Data = equipmentData{
		OrgID:             id,
		Categories:        cats,
		Inventories:       items,
		Filtered:          query != "" || category != "",
		Query:             query,
		Category:          category,
		Sort:              sort,
		ShowArchived:      showArchived,
		PageBaseURL:       template.URL(equipmentPageURL(id, query, category, sort, showArchived)),  // #nosec G203
		PrintURL:          template.URL(equipmentPrintURL(id, query, category, showArchived)),       // #nosec G203
		ToggleArchivedURL: template.URL(equipmentPageURL(id, query, category, sort, !showArchived)), // #nosec G203
		Pagination:        meta,
	}
	return app.html.Render(w, r, http.StatusOK, pages.Equipment, data)
}

func (app *application) getEquipmentNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("org_id")

	deps, appErr := app.loadEquipmentFormDeps(r, id)
	if appErr != nil {
		return appErr
	}

	data := app.html.TemplateData(r)
	data.Form = &equipment.NewForm{}
	data.Data = equipmentItemData{OrgID: id, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency}
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
		deps, appErr := app.loadEquipmentFormDeps(r, id)
		if appErr != nil {
			return appErr
		}
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentItemData{OrgID: id, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentNew, data)
	}

	if !form.Validate() {
		return reRender(&form)
	}

	if appErr := app.resolveNewFormRefs(r, id, &form); appErr != nil {
		return appErr
	}

	itemID := ksuid.New().String()
	base := equipment.Base{
		ID:             itemID,
		OrgID:          id,
		UsageTypeID:    equipment.ParseUsage(form.UsageTypeID).ID(),
		Name:           form.Name,
		CategoryID:     form.CategoryID,
		ManufacturerID: form.ManufacturerID,
		LocationID:     database.StringOrNil(form.LocationID),
		HasContent:     form.HasContentBool(),
		PurchasePrice:  form.PurchasePriceCents(),
		RentalPrice:    form.RentalPriceCents(),
		Notes:          form.Notes,
	}

	eqType := equipment.TypeFromString(form.TypeID)
	if eqType == equipment.Bulk {
		base.HasContent = false
	}
	eq, err := app.services.equipment.Create(ctx, equipment.CreateEquipment{
		Type:       eqType,
		Base:       base,
		TotalStock: form.CountInt64(),
		UnitCount:  form.CountInt64(),
	})
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to create inventory item.", Code: http.StatusInternalServerError}
	}

	if appErr := app.processEquipmentImage(r, id, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	if eq.Type == equipment.Bulk {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	}
	return nil
}

func (app *application) resolveNewFormRefs(r *http.Request, orgID string, form *equipment.NewForm) *httperr.Error {
	ctx := r.Context()
	if form.CategoryID == "" && form.CategoryName != "" {
		catID, err := app.services.equipmentcategories.EnsureByName(ctx, orgID, form.CategoryName)
		if err != nil {
			return &httperr.Error{Error: err, Message: "Failed to resolve category.", Code: http.StatusInternalServerError}
		}
		form.CategoryID = catID
	}
	if form.ManufacturerID == "" && form.ManufacturerName != "" {
		mfrID, err := app.services.manufacturers.EnsureByName(ctx, orgID, form.ManufacturerName)
		if err != nil {
			return &httperr.Error{Error: err, Message: "Failed to resolve manufacturer.", Code: http.StatusInternalServerError}
		}
		form.ManufacturerID = mfrID
	}
	if form.LocationID == "" && form.LocationName != "" {
		locID, err := app.services.locations.EnsureByName(ctx, orgID, form.LocationName)
		if err != nil {
			return &httperr.Error{Error: err, Message: "Failed to resolve location.", Code: http.StatusInternalServerError}
		}
		form.LocationID = locID
	}
	return nil
}

func (app *application) getEquipmentItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	deps, appErr := app.loadEquipmentFormDeps(r, orgID)
	if appErr != nil {
		return appErr
	}

	partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
	if loadPartOfErr != nil {
		return loadPartOfErr
	}

	data := app.html.TemplateData(r)
	data.Form = &equipment.DetailsForm{}
	itemData := equipmentItemData{OrgID: orgID, Item: item, ID: itemID, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency, PartOf: partOf, ActiveTab: "details"}
	data.Data = itemData
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
		app.resolveItemURLs(item)
		deps, appErr := app.loadEquipmentFormDeps(r, orgID)
		if appErr != nil {
			return appErr
		}
		partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
		if loadPartOfErr != nil {
			return loadPartOfErr
		}
		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency, PartOf: partOf, ActiveTab: "details"}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentDetail, data)
	}

	itemType := equipment.TypeFromString(form.TypeID)
	if err := app.services.equipment.UpdateDetails(ctx, equipment.UpdateEquipmentDetails{
		ID:             itemID,
		Type:           itemType,
		Name:           form.Name,
		CategoryID:     form.CategoryID,
		ManufacturerID: form.ManufacturerID,
		LocationID:     form.LocationID,
		Notes:          form.Notes,
		TotalStock:     form.TotalStockInt64(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID), http.StatusSeeOther)
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

	if err := app.services.equipment.UpdateProperties(ctx, equipment.UpdateEquipmentProperties{
		ID:        itemID,
		WeightG:   form.WeightGInt64(),
		WidthMM:   form.WidthMMInt64(),
		HeightMM:  form.HeightMMInt64(),
		DepthMM:   form.DepthMMInt64(),
		VoltageV:  form.VoltageVInt64(),
		PowerMW:   form.PowerMW(),
		CurrentMA: form.CurrentMA(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/properties", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentItemPricing(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	deps, appErr := app.loadEquipmentFormDeps(r, orgID)
	if appErr != nil {
		return appErr
	}

	partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
	if loadPartOfErr != nil {
		return loadPartOfErr
	}

	data := app.html.TemplateData(r)
	data.Form = &equipment.PricingForm{}
	data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency, PartOf: partOf, ActiveTab: "pricing"}
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
		app.resolveItemURLs(item)
		deps, appErr := app.loadEquipmentFormDeps(r, orgID)
		if appErr != nil {
			return appErr
		}
		partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
		if loadPartOfErr != nil {
			return loadPartOfErr
		}
		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency, PartOf: partOf, ActiveTab: "pricing"}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentPricing, data)
	}

	if err := app.services.equipment.UpdatePricing(ctx, equipment.UpdateEquipmentPricing{
		ID:            itemID,
		PurchasePrice: form.PurchasePriceCents(),
		RentalPrice:   form.RentalPriceCents(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update inventory item.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/pricing", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentItemProperties(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	deps, appErr := app.loadEquipmentFormDeps(r, orgID)
	if appErr != nil {
		return appErr
	}

	partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
	if loadPartOfErr != nil {
		return loadPartOfErr
	}

	data := app.html.TemplateData(r)
	data.Form = &equipment.PropertiesForm{}
	data.Data = equipmentItemData{OrgID: orgID, Item: item, ID: itemID, Categories: deps.Categories, Manufacturers: deps.Manufacturers, Locations: deps.Locations, Currency: deps.Currency, PartOf: partOf, ActiveTab: "properties"}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentProperties, data)
}

func (app *application) postEquipmentAddUnit(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseUnit(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	if !form.Validate() {
		return app.renderEquipmentUnits(w, r, orgID, itemID)
	}

	serialNumber := form.ManufacturerSerialNumber
	if serialNumber == "" {
		serialNumber = serial.New()
	}
	if _, err := app.services.equipment.AddUnit(ctx, equipment.AddUnit{
		ID:           ksuid.New().String(),
		OrgID:        orgID,
		EquipmentID:  itemID,
		SerialNumber: serialNumber,
		Code:         form.Code,
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
		StatusID:                 form.StatusIDInt64(),
		ManufacturerSerialNumber: form.ManufacturerSerialNumber,
		Code:                     form.Code,
		Notes:                    form.Notes,
		PurchasePrice:            form.PurchasePriceCents(),
		PurchasedAt:              form.PurchasedAtUnix(),
		NextInspectionAt:         form.NextInspectionAtUnix(),
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update unit.", Code: http.StatusInternalServerError}
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

func (app *application) getEquipmentUnits(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	app.resolveItemURLs(item)

	units, err := app.services.equipment.ListUnits(ctx, itemID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve units.", Code: http.StatusInternalServerError}
	}

	deps, appErr := app.loadEquipmentFormDeps(r, orgID)
	if appErr != nil {
		return appErr
	}

	partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
	if loadPartOfErr != nil {
		return loadPartOfErr
	}

	data := app.html.TemplateData(r)
	data.Data = equipmentUnitsData{
		OrgID:      orgID,
		ID:         itemID,
		ItemName:   item.Name,
		Units:      units,
		Item:       item,
		Categories: deps.Categories,
		PartOf:     partOf,
		ActiveTab:  "units",
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentUnits, data)
}

func (app *application) getEquipmentUnitQR(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	unit, err := app.services.equipment.GetUnit(ctx, unitID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Unit not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve unit.", Code: http.StatusInternalServerError}
	}

	content := unit.Code
	if content == "" {
		content = unit.SerialNumber
	}
	png, err := barcodes.QR(content)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to generate QR code.", Code: http.StatusInternalServerError}
	}

	filename := unitQRFilename(item.Name, content)
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
	record, appErr := app.storeEquipmentImage(r, orgID, itemID, file, header)
	if appErr != nil {
		return appErr
	}
	return app.linkEquipmentImage(r, itemID, record.ID)
}

// storeEquipmentImage processes an uploaded image and writes it to storage.
// It does not link the image to the inventory item; call linkEquipmentImage for that.
func (app *application) storeEquipmentImage(r *http.Request, orgID, itemID string, file multipart.File, header *multipart.FileHeader) (*storage.Object, *httperr.Error) {
	ctx := r.Context()

	result, err := imgpkg.Process(file)
	if err != nil {
		return nil, &httperr.Error{Error: err, Message: "Invalid image file.", Code: http.StatusUnprocessableEntity}
	}

	key := fmt.Sprintf("orgs/%s/equipment/%s", orgID, itemID)
	record, err := app.services.storageManager.Put(ctx, orgID, key, header.Filename, bytes.NewReader(result.Data), storage.Options{
		Size:        int64(len(result.Data)),
		ContentType: result.ContentType,
	})
	if err != nil {
		return nil, &httperr.Error{Error: err, Message: "Failed to store image.", Code: http.StatusInternalServerError}
	}

	return record, nil
}

// equipmentPageURL builds the paginated base URL for the inventory list.
func equipmentPageURL(orgID, query, category, sort string, showArchived bool) string {
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
	if showArchived {
		base += "archived=true&"
	}
	return base
}

// equipmentPrintURL builds the print URL with optional filter params.
func equipmentPrintURL(orgID, query, category string, showArchived bool) string {
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
	if showArchived {
		base += sep + "archived=true"
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

type equipmentFormDeps struct {
	Categories    []categories.EquipmentCategory
	Manufacturers []manufacturers.Manufacturer
	Locations     []locations.Location
	Currency      string
}

func (app *application) loadEquipmentFormDeps(r *http.Request, orgID string) (equipmentFormDeps, *httperr.Error) {
	ctx := r.Context()
	cats, err := app.services.equipment.ListCategories(ctx, orgID)
	if err != nil {
		return equipmentFormDeps{}, &httperr.Error{Error: err, Message: "Failed to retrieve categories.", Code: http.StatusInternalServerError}
	}
	mfrs, err := app.services.equipment.ListManufacturers(ctx, orgID)
	if err != nil {
		return equipmentFormDeps{}, &httperr.Error{Error: err, Message: "Failed to retrieve manufacturers.", Code: http.StatusInternalServerError}
	}
	locs, err := app.services.equipment.ListLocations(ctx, orgID)
	if err != nil {
		return equipmentFormDeps{}, &httperr.Error{Error: err, Message: "Failed to retrieve locations.", Code: http.StatusInternalServerError}
	}
	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, orgID)
	currency := ""
	if orgSettings != nil {
		currency = orgSettings.Currency
	}
	return equipmentFormDeps{Categories: cats, Manufacturers: mfrs, Locations: locs, Currency: currency}, nil
}

// linkEquipmentImage associates a storage object with an inventory item.
func (app *application) linkEquipmentImage(r *http.Request, itemID, storageObjectID string) *httperr.Error {
	if err := app.services.equipment.SetImage(r.Context(), equipment.SetImage{
		ID:              itemID,
		StorageObjectID: &storageObjectID,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to link image.", Code: http.StatusInternalServerError}
	}
	return nil
}

func (app *application) postArchiveEquipmentItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}

	if err := app.services.equipment.Archive(ctx, equipment.ArchiveEquipment{
		ID:         itemID,
		IsArchived: !item.IsArchived,
	}); err != nil {
		return &httperr.Error{Error: err, Message: "Failed to update archive status.", Code: http.StatusInternalServerError}
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID), http.StatusSeeOther)
	return nil
}

func (app *application) postDeleteEquipmentItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if err := app.services.equipment.Delete(ctx, itemID); err != nil {
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

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment", http.StatusSeeOther)
	return nil
}

func (app *application) loadContentPageData(ctx context.Context, orgID, itemID string) (*equipment.Equipment, []equipment.ContentItem, []equipment.Equipment, *httperr.Error) {
	item, err := app.services.equipment.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil, nil, &httperr.Error{Error: err, Message: "Equipment item not found.", Code: http.StatusNotFound}
		}
		return nil, nil, nil, &httperr.Error{Error: err, Message: "Failed to retrieve inventory item.", Code: http.StatusInternalServerError}
	}
	if item.Kind != equipment.Virtual && !item.HasContent {
		return nil, nil, nil, &httperr.Error{Error: nil, Message: "Content tab is not enabled for this item.", Code: http.StatusForbidden}
	}
	content, err := app.services.equipment.ListContent(ctx, itemID)
	if err != nil {
		return nil, nil, nil, &httperr.Error{Error: err, Message: "Failed to retrieve content.", Code: http.StatusInternalServerError}
	}
	all, err := app.services.equipment.ListAll(ctx, orgID)
	if err != nil {
		return nil, nil, nil, &httperr.Error{Error: err, Message: "Failed to retrieve equipment list.", Code: http.StatusInternalServerError}
	}
	return item, content, all, nil
}

func (app *application) getEquipmentContent(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, content, all, appErr := app.loadContentPageData(ctx, orgID, itemID)
	if appErr != nil {
		return appErr
	}
	app.resolveItemURLs(item)

	partOf, loadPartOfErr := app.loadPartOf(ctx, itemID)
	if loadPartOfErr != nil {
		return loadPartOfErr
	}

	data := app.html.TemplateData(r)
	data.Form = &equipment.ContentForm{}
	data.Data = equipmentContentData{OrgID: orgID, ID: itemID, Item: item, Content: content, AllEquipment: all, PartOf: partOf, ActiveTab: "content"}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentContent, data)
}

// equipmentByName returns the first equipment whose Name matches name, or nil.
func equipmentByName(all []equipment.Equipment, name string) *equipment.Equipment {
	for i := range all {
		if all[i].Name == name {
			return &all[i]
		}
	}
	return nil
}

// validateContentMember checks that member can be added as content of containerID.
// Returns a non-empty error message when the assignment should be rejected.
func validateContentMember(all []equipment.Equipment, name, containerID string) (string, string) {
	member := equipmentByName(all, name)
	if member == nil {
		return "", "No equipment found with that name."
	}
	if member.ID == containerID {
		return "", "An item cannot contain itself."
	}
	if member.HasContent || member.Kind == equipment.Virtual {
		return "", "Cannot add a container as content."
	}
	return member.ID, ""
}

func (app *application) postEquipmentAssignContent(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseContent(r)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Bad request.", Code: http.StatusBadRequest}
	}

	reRender := func(f *equipment.ContentForm, extraErr string) *httperr.Error {
		item, content, all, appErr := app.loadContentPageData(ctx, orgID, itemID)
		if appErr != nil {
			return appErr
		}
		app.resolveItemURLs(item)
		partOf, _ := app.loadPartOf(ctx, itemID)
		if extraErr != "" {
			f.AddError("assign", extraErr)
		}
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentContentData{OrgID: orgID, ID: itemID, Item: item, Content: content, AllEquipment: all, PartOf: partOf, ActiveTab: "content"}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentContent, data)
	}

	if !form.Validate() {
		return reRender(&form, "")
	}

	_, _, all, appErr := app.loadContentPageData(ctx, orgID, itemID)
	if appErr != nil {
		return appErr
	}
	memberID, memberErr := validateContentMember(all, form.MemberName, itemID)
	if memberErr != "" {
		return reRender(&form, memberErr)
	}

	_, err = app.services.equipment.AssignContent(ctx, equipment.AssignContent{
		ID:          ksuid.New().String(),
		EquipmentID: itemID,
		MemberID:    memberID,
		Quantity:    form.QuantityInt64(),
	})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			return reRender(&form, "This item is already assigned as content.")
		}
		return reRender(&form, "Failed to assign content item.")
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

func (app *application) getEquipmentPrint(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")

	org, err := app.services.orgs.GetByID(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve organization.", Code: http.StatusInternalServerError}
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")
	showArchived := qs.Get("archived") == "true"

	items, err := app.services.equipment.ListAll(ctx, orgID)
	if err != nil {
		return &httperr.Error{Error: err, Message: "Failed to retrieve inventory.", Code: http.StatusInternalServerError}
	}

	filtered := filterEquipmentForPrint(items, query, category, showArchived)

	app.resolveEquipmentURLs(filtered)

	orgSettings, _ := app.services.orgsettings.GetByOrgID(ctx, orgID)
	var currency string
	var vatRateDisplay string
	if orgSettings != nil {
		currency = orgSettings.Currency
		if orgSettings.VatRate > 0 {
			vatRateDisplay = fmt.Sprintf("%.4g%%", float64(orgSettings.VatRate)/100)
		}
	}

	data := app.html.TemplateData(r)
	data.Data = equipmentPrintData{
		OrgID:          orgID,
		OrgName:        org.Name,
		Inventories:    filtered,
		Query:          query,
		Category:       category,
		PrintDate:      time.Now().UTC().Format("2006-01-02"),
		TotalCount:     len(filtered),
		Currency:       currency,
		VatRateDisplay: vatRateDisplay,
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentPrint, data)
}

func filterEquipmentForPrint(items []equipment.Equipment, query, category string, showArchived bool) []equipment.Equipment {
	filtered := make([]equipment.Equipment, 0, len(items))
	for _, item := range items {
		if item.IsArchived != showArchived {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(query)) {
			continue
		}
		if category != "" && item.CategoryName != category {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}
