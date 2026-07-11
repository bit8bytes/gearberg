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
	PurchasePrice *Cents
	RentalPrice   *Cents
	validator.Validator
}

// PropertiesForm holds parsed input and validation state for the properties tab.
type PropertiesForm struct {
	Weight    *Grams       // user enters kg; stored as grams
	Width     *Millimeters // user enters cm; stored as mm
	Height    *Millimeters // user enters cm; stored as mm
	Depth     *Millimeters // user enters cm; stored as mm
	Power     *Milliwatts  // user enters W; stored as mW
	Current   *Milliamps   // user enters A; stored as mA
	Voltage   *Millivolts  // user enters and stored as whole volts
	WireGauge *WireGauge   // user enters mm²×100 integer; stored as-is
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

// ParsePricing reads the pricing-tab form fields from r.
func ParsePricing(r *http.Request) (PricingForm, error) {
	if err := r.ParseForm(); err != nil {
		return PricingForm{}, fmt.Errorf("parse form: %w", err)
	}
	return PricingForm{
		PurchasePrice: ParseCents(r.PostForm.Get("purchase_price")),
		RentalPrice:   ParseCents(r.PostForm.Get("rental_price")),
	}, nil
}

// ParseProperties reads the properties-tab form fields from r.
func ParseProperties(r *http.Request) (PropertiesForm, error) {
	if err := r.ParseForm(); err != nil {
		return PropertiesForm{}, fmt.Errorf("parse form: %w", err)
	}
	return PropertiesForm{
		Weight:    ParseGrams(r.PostForm.Get("weight_kg")),
		Width:     ParseMillimeters(r.PostForm.Get("width_cm")),
		Height:    ParseMillimeters(r.PostForm.Get("height_cm")),
		Depth:     ParseMillimeters(r.PostForm.Get("depth_cm")),
		Power:     ParseMilliwatts(r.PostForm.Get("power_w")),
		Current:   ParseMilliamps(r.PostForm.Get("current_a")),
		Voltage:   ParseVolts(r.PostForm.Get("voltage_v")),
		WireGauge: ParseWireGauge(r.PostForm.Get("wire_gauge_mm2_x100")),
	}, nil
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

// Validate checks PricingForm fields and returns true when all pass.
func (f *PricingForm) Validate() bool {
	if validator.NotNil(f.PurchasePrice) {
		f.Check(*f.PurchasePrice > 0, "purchase_price", "Must be greater than 0")
	}
	if validator.NotNil(f.RentalPrice) {
		f.Check(*f.RentalPrice > 0, "rental_price", "Must be greater than 0")
	}
	return f.Valid()
}

// Validate checks PropertiesForm fields and returns true when all pass.
// All fields are optional; when provided they must be greater than 0.
func (f *PropertiesForm) Validate() bool {
	if validator.NotNil(f.Weight) {
		f.Check(*f.Weight > 0, "weight_kg", "Must be greater than 0")
	}
	if validator.NotNil(f.Width) {
		f.Check(*f.Width > 0, "width_cm", "Must be greater than 0")
	}
	if validator.NotNil(f.Height) {
		f.Check(*f.Height > 0, "height_cm", "Must be greater than 0")
	}
	if validator.NotNil(f.Depth) {
		f.Check(*f.Depth > 0, "depth_cm", "Must be greater than 0")
	}
	if validator.NotNil(f.Power) {
		f.Check(*f.Power > 0, "power_w", "Must be greater than 0")
	}
	if validator.NotNil(f.Current) {
		f.Check(*f.Current > 0, "current_a", "Must be greater than 0")
	}
	if validator.NotNil(f.Voltage) {
		f.Check(*f.Voltage > 0, "voltage_v", "Must be greater than 0")
	}
	if validator.NotNil(f.WireGauge) {
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
		PurchasePrice: f.PurchasePrice,
		RentalPrice:   f.RentalPrice,
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
		PurchasePrice: e.Pricing.PurchasePrice,
		RentalPrice:   e.Pricing.RentalPrice,
	}
}

// PropertiesFormFromEquipment pre-populates a PropertiesForm from an existing Equipment's stored values.
func PropertiesFormFromEquipment(e *Equipment) PropertiesForm {
	return PropertiesForm{
		Weight:    e.Properties.Weight,
		Width:     e.Properties.Width,
		Height:    e.Properties.Height,
		Depth:     e.Properties.Depth,
		Power:     e.Properties.Power,
		Current:   e.Properties.Current,
		Voltage:   e.Properties.Voltage,
		WireGauge: e.Properties.WireGauge,
	}
}

// UnitForm holds parsed input and validation state for adding or updating a serialized unit.
type UnitForm struct {
	Code                     string
	ManufacturerSerialNumber string
	Notes                    string
	Quantity                 int64
	PurchasePrice            *Cents
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
	qty := ParseQuantity(r.PostForm.Get("quantity"))
	if qty <= 0 {
		qty = 1
	}
	return UnitForm{
		Code:                     strings.TrimSpace(r.PostForm.Get("code")),
		ManufacturerSerialNumber: strings.TrimSpace(r.PostForm.Get("serial_number")),
		Notes:                    strings.TrimSpace(r.PostForm.Get("notes")),
		Quantity:                 qty,
		PurchasePrice:            ParseCents(r.PostForm.Get("purchase_price")),
		PurchasedAt:              parseUnixDate(r.PostForm.Get("purchased_at")),
		NextInspectionAt:         parseUnixDate(r.PostForm.Get("next_inspection_at")),
		StatusID:                 parseCheckbox(r.PostForm.Get("status_id")),
	}, nil
}

// Validate checks UnitForm fields. All fields are optional.
func (f *UnitForm) Validate() bool { return f.Valid() }

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
	HasContent       string // "1" when checked, "" when unchecked
	PurchasePrice    *Cents
	RentalPrice      *Cents
	Notes            string
	Weight           *Grams
	Width            *Millimeters
	Height           *Millimeters
	Depth            *Millimeters
	Power            *Milliwatts
	Current          *Milliamps
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
		HasContent:       r.PostForm.Get("has_content"),
		PurchasePrice:    ParseCents(r.PostForm.Get("purchase_price")),
		RentalPrice:      ParseCents(r.PostForm.Get("rental_price")),
		Notes:            strings.TrimSpace(r.PostForm.Get("notes")),
		Weight:           ParseGrams(r.PostForm.Get("weight_g")),
		Width:            ParseMillimeters(r.PostForm.Get("width_mm")),
		Height:           ParseMillimeters(r.PostForm.Get("height_mm")),
		Depth:            ParseMillimeters(r.PostForm.Get("depth_mm")),
		Power:            ParseMilliwatts(r.PostForm.Get("power_w")),
		Current:          ParseMilliamps(r.PostForm.Get("current_a")),
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

// HasContentBool returns true when the has_content checkbox was checked.
func (f *NewForm) HasContentBool() bool { return f.HasContent == "1" }

// ToProperties converts the form's parsed values into a Properties sub-struct.
// Note: NewForm only captures a subset of properties (no Voltage or WireGauge).
func (f *NewForm) ToProperties() Properties {
	return Properties{
		Weight:  f.Weight,
		Width:   f.Width,
		Height:  f.Height,
		Depth:   f.Depth,
		Power:   f.Power,
		Current: f.Current,
	}
}

// ToPricing converts the form's parsed values into a Pricing sub-struct.
func (f *NewForm) ToPricing() Pricing {
	return Pricing{
		PurchasePrice: f.PurchasePrice,
		RentalPrice:   f.RentalPrice,
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
