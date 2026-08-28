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

// Package equipment provides equipment functionality.
package equipment

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/bit8bytes/gearberg/internal/equipment/tracking"
	"github.com/bit8bytes/gearberg/internal/equipment/usage"
	"github.com/bit8bytes/gearberg/internal/units"
	"github.com/bit8bytes/toolbox/validator"
)

const maxUploadBytes = 10 << 20 // 10 MiB

// DetailsForm holds parsed input and validation state for the details tab.
type DetailsForm struct {
	Type             tracking.Type
	Name             string
	CategoryID       string
	CategoryName     string // set when user typed a new category name not yet in the DB
	ManufacturerID   string
	ManufacturerName string // set when user typed a new manufacturer name not yet in the DB
	LocationID       string
	LocationName     string       // set when user typed a new location name not yet in the DB
	TotalStock       int64        // only used for bulk items; 0 means blank/unprovided
	PurchasePrice    *units.Cents // only used for bulk items
	Notes            string
	Image            multipart.File
	ImageHeader      *multipart.FileHeader
	validator.Validator
}

// PricingForm holds parsed input and validation state for the pricing tab.
type PricingForm struct {
	ResalePrice *units.Cents
	RentalPrice *units.Cents
	validator.Validator
}

// PropertiesForm holds parsed input and validation state for the properties tab.
type PropertiesForm struct {
	Weight    *units.Grams
	Width     *units.Millimeters
	Height    *units.Millimeters
	Depth     *units.Millimeters
	Power     *units.Milliwatts
	Current   *units.Milliamps
	Voltage   *units.Millivolts
	WireGauge *units.WireGauge
	validator.Validator
}

// ParseDetails reads the details-tab form fields from r, including an optional image upload.
func ParseDetails(r *http.Request) (DetailsForm, error) {
	f := DetailsForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.Type = tracking.Parse(strings.TrimSpace(r.PostForm.Get("type_id")))
	f.Name = strings.TrimSpace(r.PostForm.Get("name"))
	f.CategoryID = strings.TrimSpace(r.PostForm.Get("category_id"))
	f.CategoryName = strings.TrimSpace(r.PostForm.Get("category_name"))
	f.ManufacturerID = strings.TrimSpace(r.PostForm.Get("manufacturer_id"))
	f.ManufacturerName = strings.TrimSpace(r.PostForm.Get("manufacturer_name"))
	f.LocationID = strings.TrimSpace(r.PostForm.Get("location_id"))
	f.LocationName = strings.TrimSpace(r.PostForm.Get("location_name"))
	f.TotalStock = ParseQuantity(r.PostForm.Get("total_stock"))
	f.PurchasePrice = units.ParseCents(r.PostForm.Get("purchase_price"))
	f.Notes = strings.TrimSpace(r.PostForm.Get("notes"))
	file, header, err := r.FormFile("image")
	if err == nil {
		f.Image = file
		f.ImageHeader = header
	}
	return f, nil
}

// Validate checks DetailsForm fields. For bulk items TotalStock is required.
func (f *DetailsForm) Validate() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")

	if f.Type == tracking.Bulk {
		f.Check(f.TotalStock >= 1, "total_stock", "Must be at least 1")
		if f.PurchasePrice != nil {
			f.Check(*f.PurchasePrice > 0, "purchase_price", "Must be greater than 0")
		}
	}

	return f.Valid()
}

// ParsePricing reads the pricing-tab form fields from r.
func ParsePricing(r *http.Request) (PricingForm, error) {
	f := PricingForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.ResalePrice = units.ParseCents(r.PostForm.Get("resale_price"))
	f.RentalPrice = units.ParseCents(r.PostForm.Get("rental_price"))
	return f, nil
}

// Validate checks PricingForm fields and returns true when all pass.
func (f *PricingForm) Validate() bool {
	if f.ResalePrice != nil {
		f.Check(*f.ResalePrice > 0, "resale_price", "Must be greater than 0")
	}
	if f.RentalPrice != nil {
		f.Check(*f.RentalPrice > 0, "rental_price", "Must be greater than 0")
	}
	return f.Valid()
}

// ParseProperties reads the properties-tab form fields from r.
func ParseProperties(r *http.Request) (PropertiesForm, error) {
	f := PropertiesForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.Weight = units.ParseGrams(r.PostForm.Get("weight_kg"))
	f.Width = units.ParseMillimeters(r.PostForm.Get("width_cm"))
	f.Height = units.ParseMillimeters(r.PostForm.Get("height_cm"))
	f.Depth = units.ParseMillimeters(r.PostForm.Get("depth_cm"))
	f.Power = units.ParseMilliwatts(r.PostForm.Get("power_w"))
	f.Current = units.ParseMilliamps(r.PostForm.Get("current_a"))
	f.Voltage = units.ParseVolts(r.PostForm.Get("voltage_v"))
	f.WireGauge = units.ParseWireGauge(r.PostForm.Get("wire_gauge_mm2_x100"))
	return f, nil
}

// Validate checks PropertiesForm fields and returns true when all pass.
// All fields are optional; when provided they must be greater than 0.
// The high cyclomatic complexity is mechanical repetition across 8 independent fields, not branching logic.
//
//nolint:cyclop
func (f *PropertiesForm) Validate() bool {
	if f.Weight != nil {
		f.Check(*f.Weight > 0, "weight_kg", "Must be greater than 0")
	}
	if f.Width != nil {
		f.Check(*f.Width > 0, "width_cm", "Must be greater than 0")
	}
	if f.Height != nil {
		f.Check(*f.Height > 0, "height_cm", "Must be greater than 0")
	}
	if f.Depth != nil {
		f.Check(*f.Depth > 0, "depth_cm", "Must be greater than 0")
	}
	if f.Power != nil {
		f.Check(*f.Power > 0, "power_w", "Must be greater than 0")
	}
	if f.Current != nil {
		f.Check(*f.Current > 0, "current_a", "Must be greater than 0")
	}
	if f.Voltage != nil {
		f.Check(*f.Voltage > 0, "voltage_v", "Must be greater than 0")
	}
	if f.WireGauge != nil {
		f.Check(*f.WireGauge > 0, "wire_gauge_mm2_x100", "Must be greater than 0")
	}
	return f.Valid()
}

// ToProperties converts the form's parsed values into a Properties sub-struct.
func (f *PropertiesForm) ToProperties() Properties {
	return Properties{
		Weight:    f.Weight,
		Width:     f.Width,
		Height:    f.Height,
		Depth:     f.Depth,
		Power:     f.Power,
		Current:   f.Current,
		Voltage:   f.Voltage,
		WireGauge: f.WireGauge,
	}
}

// ToPricing converts the form's parsed values into a Pricing sub-struct.
func (f *PricingForm) ToPricing() Pricing {
	return Pricing{
		ResalePrice: f.ResalePrice,
		RentalPrice: f.RentalPrice,
	}
}

// DetailsForm pre-populates a DetailsForm from the equipment's stored values.
func (e *Equipment) DetailsForm() DetailsForm {
	f := DetailsForm{
		Type:           e.TrackingType,
		Name:           e.Name,
		CategoryID:     e.CategoryID,
		ManufacturerID: e.ManufacturerID,
		LocationID:     e.LocationID,
		TotalStock:     e.TotalStock,
		PurchasePrice:  e.BulkPurchasePrice,
		Notes:          e.Notes,
	}
	f.Errors = make(map[string]string)
	return f
}

// PricingForm pre-populates a PricingForm from the equipment's stored values.
func (e *Equipment) PricingForm() PricingForm {
	f := PricingForm{
		ResalePrice: e.Pricing.ResalePrice,
		RentalPrice: e.Pricing.RentalPrice,
	}
	f.Errors = make(map[string]string)
	return f
}

// PropertiesForm pre-populates a PropertiesForm from the equipment's stored values.
func (e *Equipment) PropertiesForm() PropertiesForm {
	f := PropertiesForm{
		Weight:    e.Properties.Weight,
		Width:     e.Properties.Width,
		Height:    e.Properties.Height,
		Depth:     e.Properties.Depth,
		Power:     e.Properties.Power,
		Current:   e.Properties.Current,
		Voltage:   e.Properties.Voltage,
		WireGauge: e.Properties.WireGauge,
	}
	f.Errors = make(map[string]string)
	return f
}

// UnitForm holds parsed input and validation state for adding or updating a serialized unit.
type UnitForm struct {
	SerialNumber             string
	ManufacturerSerialNumber string
	Remark                   string
	Quantity                 int64
	PurchasePrice            *units.Cents
	PurchasedAt              *int64 // Unix timestamp
	NextInspectionAt         *int64 // Unix timestamp
	StatusID                 int64
	validator.Validator
}

// ParseUnit reads unit form fields from r.
func ParseUnit(r *http.Request) (UnitForm, error) {
	f := UnitForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.SerialNumber = strings.TrimSpace(r.PostForm.Get("unit_serial_number"))
	f.ManufacturerSerialNumber = strings.TrimSpace(r.PostForm.Get("serial_number"))
	f.Remark = strings.TrimSpace(r.PostForm.Get("remark"))
	f.Quantity = ParseQuantity(r.PostForm.Get("quantity"))
	f.PurchasePrice = units.ParseCents(r.PostForm.Get("purchase_price"))
	f.PurchasedAt = ParseDate(r.PostForm.Get("purchased_at"))
	f.NextInspectionAt = ParseDate(r.PostForm.Get("next_inspection_at"))
	if strings.TrimSpace(r.PostForm.Get("status_id")) != "" {
		f.StatusID = 1
	} else {
		f.StatusID = 0
	}
	return f, nil
}

// Validate checks UnitForm fields. SerialNumber is required; all other fields are optional.
func (f *UnitForm) Validate() bool {
	f.Check(validator.NotBlank(f.SerialNumber), "unit_serial_number", "Serial number is required")
	f.Check(validator.MaxChars(f.SerialNumber, 200), "unit_serial_number", "Serial number cannot exceed 200 characters")
	return f.Valid()
}

// NewForm holds the parsed form input and validation state for inventory creation (both types).
type NewForm struct {
	ItemType         string // "bulk", "serialized", or "kit"
	UsageType        usage.Type
	Name             string
	CategoryID       string
	CategoryName     string // set when user typed a new category name not yet in the DB
	ManufacturerID   string
	ManufacturerName string // set when user typed a new manufacturer name not yet in the DB
	LocationID       string
	LocationName     string // set when user typed a new location name not yet in the DB
	Count            int64  // total_stock for bulk; number of units to generate for serialized/kit
	PurchasePrice    *units.Cents // bulk only; stored in equipment_bulk_items
	ResalePrice      *units.Cents
	RentalPrice      *units.Cents
	Notes            string
	Image            multipart.File
	ImageHeader      *multipart.FileHeader
	validator.Validator
}

// NewCreateForm returns a NewForm with an initialized Errors map, safe for template rendering.
func NewCreateForm() *NewForm {
	f := &NewForm{}
	f.Errors = make(map[string]string)
	return f
}

// ParseForm reads the unified create form fields from r.
func ParseForm(r *http.Request) (NewForm, error) {
	f := NewForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.ItemType = strings.TrimSpace(r.PostForm.Get("type_id"))
	f.UsageType = usage.Parse(r.PostForm.Get("usage_type_id"))
	f.Name = strings.TrimSpace(r.PostForm.Get("equipment_name"))
	f.CategoryID = strings.TrimSpace(r.PostForm.Get("category_id"))
	f.CategoryName = strings.TrimSpace(r.PostForm.Get("category_name"))
	f.ManufacturerID = strings.TrimSpace(r.PostForm.Get("manufacturer_id"))
	f.ManufacturerName = strings.TrimSpace(r.PostForm.Get("manufacturer_name"))
	f.LocationID = strings.TrimSpace(r.PostForm.Get("location_id"))
	f.LocationName = strings.TrimSpace(r.PostForm.Get("location_name"))
	f.Count = ParseQuantity(r.PostForm.Get("count"))
	f.PurchasePrice = units.ParseCents(r.PostForm.Get("purchase_price"))
	f.ResalePrice = units.ParseCents(r.PostForm.Get("resale_price"))
	f.RentalPrice = units.ParseCents(r.PostForm.Get("rental_price"))
	f.Notes = strings.TrimSpace(r.PostForm.Get("notes"))
	file, header, err := r.FormFile("image")
	if err == nil {
		f.Image = file
		f.ImageHeader = header
	}
	return f, nil
}

// Validate checks NewForm fields and returns true when all checks pass.
func (f *NewForm) Validate() bool {
	f.Check(f.ItemType == "bulk" || f.ItemType == "serialized" || f.ItemType == "kit", "type_id", "Must be bulk, serialized, or kit")
	f.Check(f.UsageType != 0, "usage_type_id", "Must be rental or sale")
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")
	f.Check(f.Count >= 1, "count", "Must be at least 1")
	return f.Valid()
}

// ContentForm holds parsed input and validation state for assigning content to an equipment item.
type ContentForm struct {
	MemberName string
	Quantity   int64
	validator.Validator
}

// NewContentForm returns a ContentForm with an initialized Errors map, safe for template rendering.
func NewContentForm() *ContentForm {
	f := &ContentForm{}
	f.Errors = make(map[string]string)
	return f
}

// ParseContent reads the content assign form fields from r.
func ParseContent(r *http.Request) (ContentForm, error) {
	f := ContentForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.MemberName = strings.TrimSpace(r.PostForm.Get("member_name"))
	f.Quantity = ParseQuantity(r.PostForm.Get("quantity"))
	return f, nil
}

// Validate checks ContentForm fields and returns true when all checks pass.
func (f *ContentForm) Validate() bool {
	f.Check(validator.NotBlank(f.MemberName), "member_name", "An item must be selected")
	f.Check(f.Quantity >= 1, "quantity", "Must be at least 1")
	return f.Valid()
}
