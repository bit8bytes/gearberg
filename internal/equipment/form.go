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

	"github.com/bit8bytes/toolbox/validator"
)

const maxUploadBytes = 10 << 20 // 10 MiB

// DetailsForm holds parsed input and validation state for the details tab.
type DetailsForm struct {
	TypeID           string // hidden field: "bulk" or "serialized"
	Name             string
	CategoryID       string
	CategoryName     string // set when user typed a new category name not yet in the DB
	ManufacturerID   string
	ManufacturerName string // set when user typed a new manufacturer name not yet in the DB
	LocationID       string
	LocationName     string // set when user typed a new location name not yet in the DB
	TotalStock       int64  // only used for bulk items; 0 means blank/unprovided
	Notes            string
	Image            multipart.File
	ImageHeader      *multipart.FileHeader
	validator.Validator
}

// PricingForm holds parsed input and validation state for the pricing tab.
type PricingForm struct {
	PurchasePrice string
	RentalPrice   string
	validator.Validator
}

// PropertiesForm holds parsed input and validation state for the properties tab.
type PropertiesForm struct {
	Weight    string
	Width     string
	Height    string
	Depth     string
	Power     string
	Current   string
	Voltage   string
	WireGauge string
	validator.Validator
}

// ParseDetails reads the details-tab form fields from r, including an optional image upload.
func ParseDetails(r *http.Request) (DetailsForm, error) {
	f := DetailsForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.TypeID = strings.TrimSpace(r.PostForm.Get("type_id"))
	f.Name = strings.TrimSpace(r.PostForm.Get("name"))
	f.CategoryID = strings.TrimSpace(r.PostForm.Get("category_id"))
	f.CategoryName = strings.TrimSpace(r.PostForm.Get("category_name"))
	f.ManufacturerID = strings.TrimSpace(r.PostForm.Get("manufacturer_id"))
	f.ManufacturerName = strings.TrimSpace(r.PostForm.Get("manufacturer_name"))
	f.LocationID = strings.TrimSpace(r.PostForm.Get("location_id"))
	f.LocationName = strings.TrimSpace(r.PostForm.Get("location_name"))
	f.TotalStock = ParseQuantity(r.PostForm.Get("total_stock"))
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

	if f.TypeID == "bulk" {
		f.Check(f.TotalStock >= 1, "total_stock", "Must be at least 1")
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
	f.PurchasePrice = strings.TrimSpace(r.PostForm.Get("purchase_price"))
	f.RentalPrice = strings.TrimSpace(r.PostForm.Get("rental_price"))
	return f, nil
}

// Validate checks PricingForm fields and returns true when all pass.
func (f *PricingForm) Validate() bool {
	if validator.NotBlank(f.PurchasePrice) {
		p := ParseCents(f.PurchasePrice)
		f.Check(p != nil, "purchase_price", "Must be a valid number")
		f.Check(p == nil || *p > 0, "purchase_price", "Must be greater than 0")
	}
	if validator.NotBlank(f.RentalPrice) {
		r := ParseCents(f.RentalPrice)
		f.Check(r != nil, "rental_price", "Must be a valid number")
		f.Check(r == nil || *r > 0, "rental_price", "Must be greater than 0")
	}
	return f.Valid()
}

// ParsedPurchasePrice parses the purchase price string to Cents. Call only after Validate().
func (f *PricingForm) ParsedPurchasePrice() *Cents { return ParseCents(f.PurchasePrice) }

// ParsedRentalPrice parses the rental price string to Cents. Call only after Validate().
func (f *PricingForm) ParsedRentalPrice() *Cents { return ParseCents(f.RentalPrice) }

// ParseProperties reads the properties-tab form fields from r.
func ParseProperties(r *http.Request) (PropertiesForm, error) {
	f := PropertiesForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.Weight = strings.TrimSpace(r.PostForm.Get("weight_kg"))
	f.Width = strings.TrimSpace(r.PostForm.Get("width_cm"))
	f.Height = strings.TrimSpace(r.PostForm.Get("height_cm"))
	f.Depth = strings.TrimSpace(r.PostForm.Get("depth_cm"))
	f.Power = strings.TrimSpace(r.PostForm.Get("power_w"))
	f.Current = strings.TrimSpace(r.PostForm.Get("current_a"))
	f.Voltage = strings.TrimSpace(r.PostForm.Get("voltage_v"))
	f.WireGauge = strings.TrimSpace(r.PostForm.Get("wire_gauge_mm2_x100"))
	return f, nil
}

// Validate checks PropertiesForm fields and returns true when all pass.
// All fields are optional; when provided they must be valid numbers greater than 0.
// The high cyclomatic complexity is mechanical repetition across 8 independent fields, not branching logic.
//
//nolint:cyclop
func (f *PropertiesForm) Validate() bool {
	if validator.NotBlank(f.Weight) {
		v := ParseGrams(f.Weight)
		f.Check(v != nil, "weight_kg", "Must be a valid number")
		f.Check(v == nil || *v > 0, "weight_kg", "Must be greater than 0")
	}
	if validator.NotBlank(f.Width) {
		v := ParseMillimeters(f.Width)
		f.Check(v != nil, "width_cm", "Must be a valid number")
		f.Check(v == nil || *v > 0, "width_cm", "Must be greater than 0")
	}
	if validator.NotBlank(f.Height) {
		v := ParseMillimeters(f.Height)
		f.Check(v != nil, "height_cm", "Must be a valid number")
		f.Check(v == nil || *v > 0, "height_cm", "Must be greater than 0")
	}
	if validator.NotBlank(f.Depth) {
		v := ParseMillimeters(f.Depth)
		f.Check(v != nil, "depth_cm", "Must be a valid number")
		f.Check(v == nil || *v > 0, "depth_cm", "Must be greater than 0")
	}
	if validator.NotBlank(f.Power) {
		v := ParseMilliwatts(f.Power)
		f.Check(v != nil, "power_w", "Must be a valid number")
		f.Check(v == nil || *v > 0, "power_w", "Must be greater than 0")
	}
	if validator.NotBlank(f.Current) {
		v := ParseMilliamps(f.Current)
		f.Check(v != nil, "current_a", "Must be a valid number")
		f.Check(v == nil || *v > 0, "current_a", "Must be greater than 0")
	}
	if validator.NotBlank(f.Voltage) {
		v := ParseVolts(f.Voltage)
		f.Check(v != nil, "voltage_v", "Must be a valid number")
		f.Check(v == nil || *v > 0, "voltage_v", "Must be greater than 0")
	}
	if validator.NotBlank(f.WireGauge) {
		v := ParseWireGauge(f.WireGauge)
		f.Check(v != nil, "wire_gauge_mm2_x100", "Must be a valid number")
		f.Check(v == nil || *v > 0, "wire_gauge_mm2_x100", "Must be greater than 0")
	}
	return f.Valid()
}

// ToProperties converts the form's parsed values into a Properties sub-struct.
func (f *PropertiesForm) ToProperties() Properties {
	return Properties{
		Weight:    ParseGrams(f.Weight),
		Width:     ParseMillimeters(f.Width),
		Height:    ParseMillimeters(f.Height),
		Depth:     ParseMillimeters(f.Depth),
		Power:     ParseMilliwatts(f.Power),
		Current:   ParseMilliamps(f.Current),
		Voltage:   ParseVolts(f.Voltage),
		WireGauge: ParseWireGauge(f.WireGauge),
	}
}

// ToPricing converts the form's parsed values into a Pricing sub-struct.
func (f *PricingForm) ToPricing() Pricing {
	return Pricing{
		PurchasePrice: f.ParsedPurchasePrice(),
		RentalPrice:   f.ParsedRentalPrice(),
	}
}

// DetailsFormFromEquipment pre-populates a DetailsForm from an existing Equipment's stored values.
func DetailsFormFromEquipment(e *Equipment) DetailsForm {
	f := DetailsForm{
		TypeID:         e.TrackingType.String(),
		Name:           e.Name,
		CategoryID:     e.CategoryID,
		ManufacturerID: e.ManufacturerID,
		LocationID:     e.LocationID,
		TotalStock:     e.TotalStock,
		Notes:          e.Notes,
	}
	f.Errors = make(map[string]string)
	return f
}

// PricingFormFromEquipment pre-populates a PricingForm from an existing Equipment's stored values.
func PricingFormFromEquipment(e *Equipment) PricingForm {
	f := PricingForm{
		PurchasePrice: e.Pricing.PurchasePrice.ToDecimal(),
		RentalPrice:   e.Pricing.RentalPrice.ToDecimal(),
	}
	f.Errors = make(map[string]string)
	return f
}

// PropertiesFormFromEquipment pre-populates a PropertiesForm from an existing Equipment's stored values.
func PropertiesFormFromEquipment(e *Equipment) PropertiesForm {
	f := PropertiesForm{
		Weight:    e.Properties.Weight.ToKG(),
		Width:     e.Properties.Width.ToCM(),
		Height:    e.Properties.Height.ToCM(),
		Depth:     e.Properties.Depth.ToCM(),
		Power:     e.Properties.Power.ToW(),
		Current:   e.Properties.Current.ToA(),
		Voltage:   e.Properties.Voltage.ToV(),
		WireGauge: e.Properties.WireGauge.String(),
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
	PurchasePrice            string
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
	f.PurchasePrice = strings.TrimSpace(r.PostForm.Get("purchase_price"))
	f.PurchasedAt = ParseDate(r.PostForm.Get("purchased_at"))
	f.NextInspectionAt = ParseDate(r.PostForm.Get("next_inspection_at"))
	f.StatusID = parseCheckbox(r.PostForm.Get("status_id"))
	return f, nil
}

// Validate checks UnitForm fields. SerialNumber is required; all other fields are optional.
func (f *UnitForm) Validate() bool {
	f.Check(validator.NotBlank(f.SerialNumber), "unit_serial_number", "Serial number is required")
	f.Check(validator.MaxChars(f.SerialNumber, 200), "unit_serial_number", "Serial number cannot exceed 200 characters")
	return f.Valid()
}

// ParsedPurchasePrice parses the purchase price string to Cents. Call only after Validate().
func (f *UnitForm) ParsedPurchasePrice() *Cents { return ParseCents(f.PurchasePrice) }

// NewForm holds the parsed form input and validation state for inventory creation (both types).
type NewForm struct {
	TypeID           string // "bulk", "serialized", or "kit"
	UsageTypeID      string // "rental" or "sale"
	Name             string
	CategoryID       string
	CategoryName     string // set when user typed a new category name not yet in the DB
	ManufacturerID   string
	ManufacturerName string // set when user typed a new manufacturer name not yet in the DB
	LocationID       string
	LocationName     string // set when user typed a new location name not yet in the DB
	Count            int64  // total_stock for bulk; number of units to generate for serialized/kit
	PurchasePrice    string
	RentalPrice      string
	Notes            string
	Image            multipart.File
	ImageHeader      *multipart.FileHeader
	validator.Validator
}

// NewNewForm returns a NewForm with an initialized Errors map, safe for template rendering.
func NewNewForm() *NewForm {
	f := &NewForm{}
	f.Errors = make(map[string]string)
	return f
}

// ParseNew reads the unified create form fields from r.
func ParseNew(r *http.Request) (NewForm, error) {
	f := NewForm{}
	f.Errors = make(map[string]string)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.TypeID = strings.TrimSpace(r.PostForm.Get("type_id"))
	f.UsageTypeID = strings.TrimSpace(r.PostForm.Get("usage_type_id"))
	f.Name = strings.TrimSpace(r.PostForm.Get("equipment_name"))
	f.CategoryID = strings.TrimSpace(r.PostForm.Get("category_id"))
	f.CategoryName = strings.TrimSpace(r.PostForm.Get("category_name"))
	f.ManufacturerID = strings.TrimSpace(r.PostForm.Get("manufacturer_id"))
	f.ManufacturerName = strings.TrimSpace(r.PostForm.Get("manufacturer_name"))
	f.LocationID = strings.TrimSpace(r.PostForm.Get("location_id"))
	f.LocationName = strings.TrimSpace(r.PostForm.Get("location_name"))
	f.Count = ParseQuantity(r.PostForm.Get("count"))
	f.PurchasePrice = strings.TrimSpace(r.PostForm.Get("purchase_price"))
	f.RentalPrice = strings.TrimSpace(r.PostForm.Get("rental_price"))
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
	f.Check(f.TypeID == "bulk" || f.TypeID == "serialized" || f.TypeID == "kit", "type_id", "Must be bulk, serialized, or kit")
	f.Check(f.UsageTypeID == "rental" || f.UsageTypeID == "sale", "usage_type_id", "Must be rental or sale")
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")
	f.Check(f.Count >= 1, "count", "Must be at least 1")
	return f.Valid()
}

// TrackingType resolves the tracking type: kit maps to Serialized, everything else is literal.
func (f *NewForm) TrackingType() TrackingType {
	if f.TypeID == "bulk" {
		return Bulk
	}
	return Serialized
}

// EquipmentType resolves the equipment type: kit maps to Kit, everything else to Standard.
func (f *NewForm) EquipmentType() Type {
	if f.TypeID == "kit" {
		return Kit
	}
	return Standard
}

// ToPricing converts the form's parsed values into a Pricing sub-struct.
func (f *NewForm) ToPricing() Pricing {
	return Pricing{
		PurchasePrice: ParseCents(f.PurchasePrice),
		RentalPrice:   ParseCents(f.RentalPrice),
	}
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
