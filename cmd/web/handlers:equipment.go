// Copyright (C) 2026 Tobias Gleiter
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bit8bytes/gearberg/internal/barcodes"
	"github.com/bit8bytes/gearberg/internal/categories"
	"github.com/bit8bytes/gearberg/internal/equipment"
	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
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
		return httperr.InternalServerError(fmt.Errorf("getEquipmentPartOfFragment: %w", err))
	}

	tmplData := app.html.TemplateData(r)
	tmplData.Data = equipmentPartOfData{
		OrgID:  orgID,
		PartOf: partOf,
	}
	return app.html.RenderFragment(w, r, http.StatusOK, fragments.EquipmentPartOf, tmplData)
}

type equipmentData struct {
	OrgID           string
	Categories      []categories.Category
	Inventories     []equipment.Equipment
	Filtered        bool
	Query           string
	Category        string
	Inspection      string
	Sort            string
	PageBaseURL     template.URL
	PrintURL        template.URL
	Pagination      pagination.Metadata
	ImportSessionID string
}

func (app *application) getEquipment(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id, err := uid.Parse(r.PathValue("org_id"))
	if err != nil {
		return httperr.BadRequest(err)
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")
	inspection := qs.Get("inspection")
	sort := qs.Get("sort")

	page, err := strconv.Atoi(qs.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	items, meta, err := app.services.equipment.List(ctx, equipment.ListParams{
		OrgID:            id,
		Query:            query,
		Category:         category,
		InspectionFilter: inspection,
		Filters:          pagination.Filters{Page: page, PageSize: 25},
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	app.resolveEquipmentURLs(items)

	cats, err := app.services.equipmentcategories.List(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	var importSessionID string
	if session, err := app.services.equipmentImports.GetStagedSession(ctx, id); err == nil {
		importSessionID = session.ID
	}

	tmpl := app.html.TemplateData(r)
	tmpl.Data = equipmentData{
		OrgID:           id,
		Categories:      cats,
		Inventories:     items,
		Filtered:        query != "" || category != "" || inspection != "",
		Query:           query,
		Category:        category,
		Inspection:      inspection,
		Sort:            sort,
		PageBaseURL:     template.URL(equipmentPageURL(id, query, category, inspection, sort)),  // #nosec G203
		PrintURL:        template.URL(equipmentPrintURL(id, query, category, inspection, sort)), // #nosec G203
		Pagination:      meta,
		ImportSessionID: importSessionID,
	}

	// HTMX live-search: return only the results fragment so the page URL
	// update (hx-push-url) reflects the current filters without a full reload.
	if htmx.IsRequest(r) {
		return app.html.RenderFragment(w, r, http.StatusOK, fragments.EquipmentSearch, tmpl)
	}
	return app.html.Render(w, r, http.StatusOK, pages.Equipment, tmpl)
}

type equipmentItemData struct {
	OrgID     string
	Item      *equipment.Equipment
	ID        string
	ActiveTab string
	Currency  money.Currency
}

func (app *application) getEquipmentNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	orgSettings, err := app.services.orgsettings.Get(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	var currency money.Currency
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	data.Form = equipment.NewCreateForm()
	data.Data = equipmentItemData{OrgID: id, Currency: currency}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentNew, data)
}

//nolint:cyclop
func (app *application) postEquipmentNew(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	id := r.PathValue("org_id")

	orgSettings, err := app.services.orgsettings.Get(ctx, id)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	var currency money.Currency
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	form, err := equipment.ParseForm(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	fail := func(f *equipment.NewForm) *httperr.Error {
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentItemData{OrgID: id, Currency: currency}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentNew, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	catID := form.CategoryID
	if catID == "" && form.CategoryName != "" {
		catID, err = app.services.equipmentcategories.Upsert(ctx, id, form.CategoryName)
		if err != nil {
			return httperr.InternalServerError(err)
		}
	}
	mfrID := form.ManufacturerID
	if mfrID == "" && form.ManufacturerName != "" {
		mfrID, err = app.services.manufacturers.Upsert(ctx, id, form.ManufacturerName)
		if err != nil {
			return httperr.InternalServerError(err)
		}
	}
	locID := form.LocationID
	if locID == "" && form.LocationName != "" {
		locID, err = app.services.locations.Upsert(ctx, id, form.LocationName)
		if err != nil {
			return httperr.InternalServerError(err)
		}
	}

	base := equipment.Base{
		OrgID:          id,
		UsageTypeID:    form.UsageType.ID(),
		Name:           form.Name,
		CategoryID:     catID,
		ManufacturerID: mfrID,
		LocationID:     locID,
		Notes:          form.Notes,
		Pricing: equipment.Pricing{
			PurchasePrice: form.PurchasePrice,
			RentalPrice:   form.RentalPrice,
		},
	}

	eq, err := app.services.equipment.Create(ctx, equipment.CreateEquipment{
		ItemType:   form.ItemType,
		Base:       base,
		TotalStock: form.Count,
		UnitCount:  form.Count,
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	if appErr := app.processEquipmentImage(r, id, eq.ID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	switch eq.TrackingType {
	case tracking.Bulk:
		http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment", http.StatusSeeOther)
	case tracking.Serialized:
		if eq.Type == equipment.KitType {
			http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment/"+url.PathEscape(eq.ID)+"/content", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/orgs/"+url.PathEscape(id)+"/equipment/"+url.PathEscape(eq.ID)+"/units", http.StatusSeeOther)
		}
	}
	return nil
}

func (app *application) getEquipmentItem(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	app.resolveItemURLs(item)

	orgSettings, err := app.services.orgsettings.Get(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}
	var currency money.Currency
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	f := item.DetailsForm()
	data.Form = &f
	data.Data = equipmentItemData{
		OrgID:     orgID,
		Item:      item,
		ID:        itemID,
		ActiveTab: "details",
		Currency:  currency,
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentDetail, data)
}

//nolint:cyclop
func (app *application) postEquipmentItemDetails(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseDetails(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	if appErr := app.processEquipmentImage(r, orgID, itemID, form.Image, form.ImageHeader); appErr != nil {
		return appErr
	}

	orgSettings, err := app.services.orgsettings.Get(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}
	var currency money.Currency
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	fail := func(f *equipment.DetailsForm) *httperr.Error {
		item, err := app.services.equipment.Get(ctx, itemID)
		if err != nil {
			return httperr.InternalServerError(err)
		}
		data := app.html.TemplateData(r)
		data.Form = f
		data.Data = equipmentItemData{
			OrgID:     orgID,
			Item:      item,
			ID:        itemID,
			ActiveTab: "details",
			Currency:  currency,
		}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentDetail, data)
	}

	if !form.Validate() {
		return fail(&form)
	}

	catID := form.CategoryID
	if catID == "" && form.CategoryName != "" {
		catID, err = app.services.equipmentcategories.Upsert(ctx, orgID, form.CategoryName)
		if err != nil {
			return httperr.InternalServerError(err)
		}
	}
	mfrID := form.ManufacturerID
	if mfrID == "" && form.ManufacturerName != "" {
		mfrID, err = app.services.manufacturers.Upsert(ctx, orgID, form.ManufacturerName)
		if err != nil {
			return httperr.InternalServerError(err)
		}
	}
	locID := form.LocationID
	if locID == "" && form.LocationName != "" {
		locID, err = app.services.locations.Upsert(ctx, orgID, form.LocationName)
		if err != nil {
			return httperr.InternalServerError(err)
		}
	}

	if err := app.services.equipment.UpdateDetails(ctx, equipment.UpdateEquipmentDetails{
		ID:             itemID,
		OrgID:          orgID,
		Type:           form.Type,
		Name:           form.Name,
		CategoryID:     catID,
		ManufacturerID: mfrID,
		LocationID:     locID,
		Notes:          form.Notes,
		TotalStock:     form.TotalStock,
		PurchasePrice:  form.PurchasePrice,
	}); err != nil {
		return httperr.InternalServerError(err)
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
		return httperr.BadRequest(err)
	}

	if !form.Validate() {
		item, err := app.services.equipment.Get(ctx, itemID)
		if err != nil {
			return httperr.InternalServerError(err)
		}

		app.resolveItemURLs(item)

		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{
			OrgID:     orgID,
			Item:      item,
			ID:        itemID,
			ActiveTab: "properties",
		}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentProperties, data)
	}

	if err := app.services.equipment.UpdateProperties(ctx, equipment.UpdateEquipmentProperties{
		ID:         itemID,
		Properties: form.ToProperties(),
	}); err != nil {
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/properties#save", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentItemPricing(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	app.resolveItemURLs(item)

	orgSettings, err := app.services.orgsettings.Get(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	var currency money.Currency
	if orgSettings != nil {
		currency = orgSettings.Currency
	}

	data := app.html.TemplateData(r)
	f := item.PricingForm()
	data.Form = &f
	data.Data = equipmentItemData{
		OrgID:     orgID,
		Item:      item,
		ID:        itemID,
		ActiveTab: "pricing",
		Currency:  currency,
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentPricing, data)
}

func (app *application) postEquipmentItemPricing(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParsePricing(r)
	if err != nil {
		return httperr.BadRequest(err)
	}

	if !form.Validate() {
		item, err := app.services.equipment.Get(ctx, itemID)
		if err != nil {
			return httperr.InternalServerError(err)
		}

		orgSettings, err := app.services.orgsettings.Get(ctx, orgID)
		if err != nil {
			return httperr.InternalServerError(err)
		}

		var currency money.Currency
		if orgSettings != nil {
			currency = orgSettings.Currency
		}

		data := app.html.TemplateData(r)
		data.Form = &form
		data.Data = equipmentItemData{
			OrgID:     orgID,
			Item:      item,
			ID:        itemID,
			ActiveTab: "pricing",
			Currency:  currency,
		}
		return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentPricing, data)
	}

	if err := app.services.equipment.UpdatePricing(ctx, equipment.UpdateEquipmentPricing{
		ID:      itemID,
		Pricing: form.ToPricing(),
	}); err != nil {
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/pricing#save", http.StatusSeeOther)
	return nil
}

func (app *application) getEquipmentItemProperties(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	item, err := app.services.equipment.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	app.resolveItemURLs(item)

	data := app.html.TemplateData(r)
	f := item.PropertiesForm()
	data.Form = &f
	data.Data = equipmentItemData{
		OrgID:     orgID,
		Item:      item,
		ID:        itemID,
		ActiveTab: "properties",
	}
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
		if errors.Is(err, equipment.ErrNotSerializedUnit) {
			return &httperr.Error{
				Error:   err,
				Message: "Units can only be added to serialized equipment.",
				Code:    http.StatusBadRequest,
			}
		}
		return httperr.InternalServerError(err)
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
		return httperr.BadRequest(err)
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
		PurchasePrice:            form.PurchasePrice,
		PurchasedAt:              form.PurchasedAt,
		NextInspectionAt:         form.NextInspectionAt,
	}); err != nil {
		if errors.Is(err, equipment.ErrConflict) {
			return app.renderEquipmentUnits(w, r, orgID, itemID)
		}
		return httperr.InternalServerError(err)
	}

	http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID)+"/units", http.StatusSeeOther)
	return nil
}

func (app *application) postEquipmentBulkUpdateInspection(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		return httperr.BadRequest(err)
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
		return httperr.InternalServerError(err)
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
		return httperr.InternalServerError(err)
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

	item, err := app.services.equipment.GetUnitsContainer(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		if errors.Is(err, equipment.ErrNoUnitsTab) {
			http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID), http.StatusSeeOther)
			return nil
		}
		return httperr.InternalServerError(err)
	}

	app.resolveItemURLs(item)

	units, err := app.services.equipment.ListUnits(ctx, itemID)
	if err != nil {
		return httperr.InternalServerError(err)
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

	item, err := app.services.equipment.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	unit, err := app.services.equipment.GetUnit(ctx, unitID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	png, err := barcodes.QR(unit.SerialNumber)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	filename := unitQRFilename(item.Name, unit.SerialNumber)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(png); err != nil { //nolint:gosec // png bytes are generated internally, not from user input
		return httperr.InternalServerError(err)
	}
	return nil
}

func (app *application) getEquipmentUnitBarcode(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	itemID := r.PathValue("id")
	unitID := r.PathValue("unit_id")

	item, err := app.services.equipment.Get(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	unit, err := app.services.equipment.GetUnit(ctx, unitID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		return httperr.InternalServerError(err)
	}

	png, err := barcodes.Code128(unit.SerialNumber)
	if err != nil {
		return httperr.InternalServerError(err)
	}

	filename := unitQRFilename(item.Name, unit.SerialNumber)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := w.Write(png); err != nil { //nolint:gosec // png bytes are generated internally, not from user input
		return httperr.InternalServerError(err)
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
		return &httperr.Error{
			Error:   err,
			Message: "Invalid image file.",
			Code:    http.StatusUnprocessableEntity,
		}
	}

	key := fmt.Sprintf("orgs/%s/equipment/%s", orgID, itemID)
	record, err := app.services.storageManager.Put(ctx, orgID, key, header.Filename, bytes.NewReader(result.Data), storage.Options{
		Size:        int64(len(result.Data)),
		ContentType: result.ContentType,
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	if err := app.services.equipment.SetImage(ctx, equipment.SetImage{
		ID:              itemID,
		StorageObjectID: &record.ID,
	}); err != nil {
		return httperr.InternalServerError(err)
	}
	return nil
}

// equipmentPageURL builds the paginated base URL for the inventory list.
func equipmentPageURL(orgID, query, category, inspection, sort string) string {
	base := "/orgs/" + url.PathEscape(orgID) + "/equipment?"
	if category != "" {
		base += "category=" + url.QueryEscape(category) + "&"
	}
	if query != "" {
		base += "q=" + url.QueryEscape(query) + "&"
	}
	if inspection != "" {
		base += "inspection=" + url.QueryEscape(inspection) + "&"
	}
	if sort != "" {
		base += "sort=" + url.QueryEscape(sort) + "&"
	}
	return base
}

// equipmentPrintURL builds the print URL with optional filter params.
func equipmentPrintURL(orgID, query, category, inspection, sort string) string {
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
	if inspection != "" {
		base += sep + "inspection=" + url.QueryEscape(inspection)
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
		return httperr.InternalServerError(err)
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
			return httperr.NotFound(err)
		}
		if errors.Is(err, equipment.ErrNoContentTab) {
			http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID), http.StatusSeeOther)
			return nil
		}
		return httperr.InternalServerError(err)
	}
	app.resolveItemURLs(item)

	content, err := app.services.equipment.ListContent(ctx, itemID)
	if err != nil {
		return httperr.InternalServerError(err)
	}
	all, _, err := app.services.equipment.List(ctx, equipment.ListParams{
		OrgID: orgID,
		Filters: pagination.Filters{
			Page:     1,
			PageSize: math.MaxInt32,
		},
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	data := app.html.TemplateData(r)
	data.Form = equipment.NewContentForm()
	data.Data = equipmentContentData{
		OrgID:        orgID,
		ID:           itemID,
		Item:         item,
		Content:      content,
		AllEquipment: all,
		ActiveTab:    "content",
	}
	return app.html.Render(w, r, http.StatusOK, pages.EquipmentContent, data)
}

func (app *application) renderEquipmentContentForm(w http.ResponseWriter, r *http.Request, orgID, itemID string, f *equipment.ContentForm, extraErr string) *httperr.Error {
	ctx := r.Context()
	item, err := app.services.equipment.GetContentContainer(ctx, itemID)
	if err != nil {
		if errors.Is(err, equipment.ErrNotFound) {
			return httperr.NotFound(err)
		}
		if errors.Is(err, equipment.ErrNoContentTab) {
			http.Redirect(w, r, "/orgs/"+url.PathEscape(orgID)+"/equipment/"+url.PathEscape(itemID), http.StatusSeeOther)
			return nil
		}
		return httperr.InternalServerError(err)
	}
	app.resolveItemURLs(item)
	content, err := app.services.equipment.ListContent(ctx, itemID)
	if err != nil {
		return httperr.InternalServerError(err)
	}
	all, _, err := app.services.equipment.List(ctx, equipment.ListParams{
		OrgID: orgID,
		Filters: pagination.Filters{
			Page:     1,
			PageSize: math.MaxInt32,
		},
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}
	if extraErr != "" {
		f.AddError("assign", extraErr)
	}
	data := app.html.TemplateData(r)
	data.Form = f
	data.Data = equipmentContentData{
		OrgID:        orgID,
		ID:           itemID,
		Item:         item,
		Content:      content,
		AllEquipment: all,
		ActiveTab:    "content",
	}
	return app.html.Render(w, r, http.StatusUnprocessableEntity, pages.EquipmentContent, data)
}

func (app *application) postEquipmentAssignContent(w http.ResponseWriter, r *http.Request) *httperr.Error {
	ctx := r.Context()
	orgID := r.PathValue("org_id")
	itemID := r.PathValue("id")

	form, err := equipment.ParseContent(r)
	if err != nil {
		return httperr.BadRequest(err)
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
		return httperr.InternalServerError(err)
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
		return httperr.InternalServerError(err)
	}

	qs := r.URL.Query()
	query := qs.Get("q")
	category := qs.Get("category")
	filtered, _, err := app.services.equipment.List(ctx, equipment.ListParams{
		OrgID:    orgID,
		Query:    query,
		Category: category,
		Filters: pagination.Filters{
			Page:     1,
			PageSize: math.MaxInt32,
		},
	})
	if err != nil {
		return httperr.InternalServerError(err)
	}

	app.resolveEquipmentURLs(filtered)

	orgSettings, err := app.services.orgsettings.Get(ctx, orgID)
	if err != nil {
		return httperr.InternalServerError(err)
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
