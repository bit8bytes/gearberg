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
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return DetailsForm{}, fmt.Errorf("parse form: %w", err)
	}
	f := DetailsForm{
		TypeID:           strings.TrimSpace(r.PostForm.Get("type_id")),
		Name:             strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID:       strings.TrimSpace(r.PostForm.Get("category_id")),
		CategoryName:     strings.TrimSpace(r.PostForm.Get("category_name")),
		ManufacturerID:   strings.TrimSpace(r.PostForm.Get("manufacturer_id")),
		ManufacturerName: strings.TrimSpace(r.PostForm.Get("manufacturer_name")),
		LocationID:       strings.TrimSpace(r.PostForm.Get("location_id")),
		LocationName:     strings.TrimSpace(r.PostForm.Get("location_name")),
		TotalStock:       ParseQuantity(r.PostForm.Get("total_stock")),
		Notes:            strings.TrimSpace(r.PostForm.Get("notes")),
	}
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
	if err := r.ParseForm(); err != nil {
		return PricingForm{}, fmt.Errorf("parse form: %w", err)
	}
	return PricingForm{
		PurchasePrice: strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:   strings.TrimSpace(r.PostForm.Get("rental_price")),
	}, nil
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
	if err := r.ParseForm(); err != nil {
		return PropertiesForm{}, fmt.Errorf("parse form: %w", err)
	}
	return PropertiesForm{
		Weight:    strings.TrimSpace(r.PostForm.Get("weight_kg")),
		Width:     strings.TrimSpace(r.PostForm.Get("width_cm")),
		Height:    strings.TrimSpace(r.PostForm.Get("height_cm")),
		Depth:     strings.TrimSpace(r.PostForm.Get("depth_cm")),
		Power:     strings.TrimSpace(r.PostForm.Get("power_w")),
		Current:   strings.TrimSpace(r.PostForm.Get("current_a")),
		Voltage:   strings.TrimSpace(r.PostForm.Get("voltage_v")),
		WireGauge: strings.TrimSpace(r.PostForm.Get("wire_gauge_mm2_x100")),
	}, nil
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
	return DetailsForm{
		TypeID:         e.Type.String(),
		Name:           e.Name,
		CategoryID:     e.CategoryID,
		ManufacturerID: e.ManufacturerID,
		LocationID:     e.LocationID,
		TotalStock:     e.TotalStock,
		Notes:          e.Notes,
	}
}

// PricingFormFromEquipment pre-populates a PricingForm from an existing Equipment's stored values.
func PricingFormFromEquipment(e *Equipment) PricingForm {
	return PricingForm{
		PurchasePrice: e.Pricing.PurchasePrice.ToDecimal(),
		RentalPrice:   e.Pricing.RentalPrice.ToDecimal(),
	}
}

// PropertiesFormFromEquipment pre-populates a PropertiesForm from an existing Equipment's stored values.
func PropertiesFormFromEquipment(e *Equipment) PropertiesForm {
	return PropertiesForm{
		Weight:    e.Properties.Weight.ToKG(),
		Width:     e.Properties.Width.ToCM(),
		Height:    e.Properties.Height.ToCM(),
		Depth:     e.Properties.Depth.ToCM(),
		Power:     e.Properties.Power.ToW(),
		Current:   e.Properties.Current.ToA(),
		Voltage:   e.Properties.Voltage.ToV(),
		WireGauge: e.Properties.WireGauge.String(),
	}
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
	if err := r.ParseForm(); err != nil {
		return UnitForm{}, fmt.Errorf("parse form: %w", err)
	}
	return UnitForm{
		SerialNumber:             strings.TrimSpace(r.PostForm.Get("unit_serial_number")),
		ManufacturerSerialNumber: strings.TrimSpace(r.PostForm.Get("serial_number")),
		Remark:                   strings.TrimSpace(r.PostForm.Get("remark")),
		Quantity:                 ParseQuantity(r.PostForm.Get("quantity")),
		PurchasePrice:            strings.TrimSpace(r.PostForm.Get("purchase_price")),
		PurchasedAt:              ParseDate(r.PostForm.Get("purchased_at")),
		NextInspectionAt:         ParseDate(r.PostForm.Get("next_inspection_at")),
		StatusID:                 parseCheckbox(r.PostForm.Get("status_id")),
	}, nil
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
	TypeID           string // "bulk" or "serialized"
	UsageTypeID      string // "rental" or "sale"
	Name             string
	CategoryID       string
	CategoryName     string // set when user typed a new category name not yet in the DB
	ManufacturerID   string
	ManufacturerName string // set when user typed a new manufacturer name not yet in the DB
	LocationID       string
	LocationName     string // set when user typed a new location name not yet in the DB
	Count            int64  // total_stock for bulk; number of units to generate for serialized
	HasContent       bool
	PurchasePrice    string
	RentalPrice      string
	Notes            string
	Image            multipart.File
	ImageHeader      *multipart.FileHeader
	validator.Validator
}

// ParseNew reads the unified create form fields from r.
func ParseNew(r *http.Request) (NewForm, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return NewForm{}, fmt.Errorf("parse form: %w", err)
	}
	f := NewForm{
		TypeID:           strings.TrimSpace(r.PostForm.Get("type_id")),
		UsageTypeID:      strings.TrimSpace(r.PostForm.Get("usage_type_id")),
		Name:             strings.TrimSpace(r.PostForm.Get("equipment_name")),
		CategoryID:       strings.TrimSpace(r.PostForm.Get("category_id")),
		CategoryName:     strings.TrimSpace(r.PostForm.Get("category_name")),
		ManufacturerID:   strings.TrimSpace(r.PostForm.Get("manufacturer_id")),
		ManufacturerName: strings.TrimSpace(r.PostForm.Get("manufacturer_name")),
		LocationID:       strings.TrimSpace(r.PostForm.Get("location_id")),
		LocationName:     strings.TrimSpace(r.PostForm.Get("location_name")),
		Count:            ParseQuantity(r.PostForm.Get("count")),
		HasContent:       r.PostForm.Get("has_content") == "1",
		PurchasePrice:    strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:      strings.TrimSpace(r.PostForm.Get("rental_price")),
		Notes:            strings.TrimSpace(r.PostForm.Get("notes")),
	}
	file, header, err := r.FormFile("image")
	if err == nil {
		f.Image = file
		f.ImageHeader = header
	}
	return f, nil
}

// Validate checks NewForm fields and returns true when all checks pass.
func (f *NewForm) Validate() bool {
	f.Check(f.TypeID == "bulk" || f.TypeID == "serialized", "type_id", "Must be bulk or serialized")
	f.Check(f.UsageTypeID == "rental" || f.UsageTypeID == "sale", "usage_type_id", "Must be rental or sale")
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")
	f.Check(f.Count >= 1, "count", "Must be at least 1")
	return f.Valid()
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

// ParseContent reads the content assign form fields from r.
func ParseContent(r *http.Request) (ContentForm, error) {
	if err := r.ParseForm(); err != nil {
		return ContentForm{}, fmt.Errorf("parse form: %w", err)
	}
	return ContentForm{
		MemberName: strings.TrimSpace(r.PostForm.Get("member_name")),
		Quantity:   ParseQuantity(r.PostForm.Get("quantity")),
	}, nil
}

// Validate checks ContentForm fields and returns true when all checks pass.
func (f *ContentForm) Validate() bool {
	f.Check(validator.NotBlank(f.MemberName), "member_name", "An item must be selected")
	f.Check(f.Quantity >= 1, "quantity", "Must be at least 1")
	return f.Valid()
}
