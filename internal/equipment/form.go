package equipment

import (
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bit8bytes/toolbox/validator"
)

const maxUploadBytes = 10 << 20 // 10 MiB

// DetailsForm holds parsed input and validation state for the details tab.
type DetailsForm struct {
	TypeID         string // hidden field: "bulk" or "serialized"
	Name           string
	CategoryID     string
	ManufacturerID string
	LocationID     string
	TotalStock     string // only used for bulk items
	Notes          string
	Image          multipart.File
	ImageHeader    *multipart.FileHeader
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
	WeightG  string
	WidthMM  string
	HeightMM string
	DepthMM  string
	VoltageV string // user enters volts; stored as volts (integer)
	PowerW   string // user enters watts; stored as milliwatts
	CurrentA string // user enters amps; stored as milliamps
	validator.Validator
}

// normalizeDecimal replaces comma with dot so both European and US number
// formats are accepted before parsing (e.g. "19,99" → "19.99").
func normalizeDecimal(s string) string {
	return strings.ReplaceAll(s, ",", ".")
}

// ParseDetails reads the details-tab form fields from r, including an optional image upload.
func ParseDetails(r *http.Request) (DetailsForm, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return DetailsForm{}, fmt.Errorf("parse form: %w", err)
	}
	f := DetailsForm{
		TypeID:         strings.TrimSpace(r.PostForm.Get("type_id")),
		Name:           strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID:     strings.TrimSpace(r.PostForm.Get("category_id")),
		ManufacturerID: strings.TrimSpace(r.PostForm.Get("manufacturer_id")),
		LocationID:     strings.TrimSpace(r.PostForm.Get("location_id")),
		TotalStock:     strings.TrimSpace(r.PostForm.Get("total_stock")),
		Notes:          strings.TrimSpace(r.PostForm.Get("notes")),
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
		PurchasePrice: strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:   strings.TrimSpace(r.PostForm.Get("rental_price")),
	}, nil
}

// ParseProperties reads the properties-tab form fields from r.
func ParseProperties(r *http.Request) (PropertiesForm, error) {
	if err := r.ParseForm(); err != nil {
		return PropertiesForm{}, fmt.Errorf("parse form: %w", err)
	}
	return PropertiesForm{
		WeightG:  strings.TrimSpace(r.PostForm.Get("weight_g")),
		WidthMM:  strings.TrimSpace(r.PostForm.Get("width_mm")),
		HeightMM: strings.TrimSpace(r.PostForm.Get("height_mm")),
		DepthMM:  strings.TrimSpace(r.PostForm.Get("depth_mm")),
		VoltageV: strings.TrimSpace(r.PostForm.Get("voltage_v")),
		PowerW:   strings.TrimSpace(r.PostForm.Get("power_w")),
		CurrentA: strings.TrimSpace(r.PostForm.Get("current_a")),
	}, nil
}

// Validate checks DetailsForm fields. For bulk items TotalStock is required.
func (f *DetailsForm) Validate() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")

	if f.TypeID == "bulk" {
		if validator.NotBlank(f.TotalStock) {
			n, err := strconv.ParseInt(f.TotalStock, 10, 64)
			f.Check(err == nil, "total_stock", "Must be a whole number")
			f.Check(err != nil || n >= 1, "total_stock", "Must be at least 1")
		} else {
			f.AddError("total_stock", "This field cannot be blank")
		}
	}

	return f.Valid()
}

// Validate checks PricingForm fields and returns true when all pass.
func (f *PricingForm) Validate() bool {
	if validator.NotBlank(f.PurchasePrice) {
		_, err := strconv.ParseFloat(normalizeDecimal(f.PurchasePrice), 64)
		f.Check(err == nil, "purchase_price", "Must be a valid number")
	}

	if validator.NotBlank(f.RentalPrice) {
		_, err := strconv.ParseFloat(normalizeDecimal(f.RentalPrice), 64)
		f.Check(err == nil, "rental_price", "Must be a valid number")
	}

	return f.Valid()
}

// Validate checks PropertiesForm fields and returns true when all pass.
// All fields are optional; this method exists for a consistent call pattern.
func (f *PropertiesForm) Validate() bool { return f.Valid() }

// InspectionForm holds parsed input and validation state for the inspection tab.
type InspectionForm struct {
	IntervalDays string
	validator.Validator
}

// ParseInspection reads the inspection-tab form fields from r.
func ParseInspection(r *http.Request) (InspectionForm, error) {
	if err := r.ParseForm(); err != nil {
		return InspectionForm{}, fmt.Errorf("parse form: %w", err)
	}
	return InspectionForm{
		IntervalDays: strings.TrimSpace(r.PostForm.Get("inspection_interval_days")),
	}, nil
}

// Validate checks InspectionForm fields and returns true when all pass.
func (f *InspectionForm) Validate() bool {
	if validator.NotBlank(f.IntervalDays) {
		n, err := strconv.ParseInt(f.IntervalDays, 10, 64)
		f.Check(err == nil, "inspection_interval_days", "Must be a whole number")
		f.Check(err != nil || n >= 1, "inspection_interval_days", "Must be at least 1")
	}
	return f.Valid()
}

// IntervalDaysInt64 returns the parsed interval as *int64 (nil when blank).
func (f *InspectionForm) IntervalDaysInt64() *int64 { return parseOptionalInt64(f.IntervalDays) }

// TotalStockInt64 returns the parsed TotalStock value. Call only after Validate() returns true.
func (f *DetailsForm) TotalStockInt64() int64 {
	n, _ := strconv.ParseInt(f.TotalStock, 10, 64)
	return n
}

// PurchasePriceCents returns the purchase price in the smallest currency unit, or nil when blank.
func (f *PricingForm) PurchasePriceCents() *int64 {
	return parseOptionalCents(f.PurchasePrice)
}

// RentalPriceCents returns the rental price in the smallest currency unit, or nil when blank.
func (f *PricingForm) RentalPriceCents() *int64 {
	return parseOptionalCents(f.RentalPrice)
}

// WeightGInt64 returns weight_g as *int64 (nil when blank).
func (f *PropertiesForm) WeightGInt64() *int64 { return parseOptionalInt64(f.WeightG) }

// WidthMMInt64 returns width_mm as *int64 (nil when blank).
func (f *PropertiesForm) WidthMMInt64() *int64 { return parseOptionalInt64(f.WidthMM) }

// HeightMMInt64 returns height_mm as *int64 (nil when blank).
func (f *PropertiesForm) HeightMMInt64() *int64 { return parseOptionalInt64(f.HeightMM) }

// DepthMMInt64 returns depth_mm as *int64 (nil when blank).
func (f *PropertiesForm) DepthMMInt64() *int64 { return parseOptionalInt64(f.DepthMM) }

// VoltageVInt64 returns voltage_v as *int64 (nil when blank).
func (f *PropertiesForm) VoltageVInt64() *int64 { return parseOptionalInt64(f.VoltageV) }

// PowerMW converts the user-entered watts string to milliwatts, or nil when blank.
func (f *PropertiesForm) PowerMW() *int64 { return parseOptionalWattsToMW(f.PowerW) }

// CurrentMA converts the user-entered amps string to milliamps, or nil when blank.
func (f *PropertiesForm) CurrentMA() *int64 { return parseOptionalAmpsToMA(f.CurrentA) }

// UnitForm holds parsed input and validation state for adding or updating a serialized unit.
type UnitForm struct {
	Code                     string
	ManufacturerSerialNumber string
	Notes                    string
	Quantity                 string
	PurchasePrice            string // decimal e.g. "19.99"
	PurchasedAt              string // YYYY-MM-DD
	NextInspectionAt         string // YYYY-MM-DD
	StatusID                 string
	validator.Validator
}

// ParseUnit reads unit form fields from r.
func ParseUnit(r *http.Request) (UnitForm, error) {
	if err := r.ParseForm(); err != nil {
		return UnitForm{}, fmt.Errorf("parse form: %w", err)
	}
	return UnitForm{
		Code:                     strings.TrimSpace(r.PostForm.Get("code")),
		ManufacturerSerialNumber: strings.TrimSpace(r.PostForm.Get("serial_number")),
		Notes:                    strings.TrimSpace(r.PostForm.Get("notes")),
		Quantity:                 strings.TrimSpace(r.PostForm.Get("quantity")),
		PurchasePrice:            strings.TrimSpace(r.PostForm.Get("purchase_price")),
		PurchasedAt:              strings.TrimSpace(r.PostForm.Get("purchased_at")),
		NextInspectionAt:         strings.TrimSpace(r.PostForm.Get("next_inspection_at")),
		StatusID:                 strings.TrimSpace(r.PostForm.Get("status_id")),
	}, nil
}

// Validate checks UnitForm fields. All fields are optional.
func (f *UnitForm) Validate() bool { return f.Valid() }

// StatusIDInt64 returns the parsed StatusID as int64. Returns 0 when blank.
func (f *UnitForm) StatusIDInt64() int64 {
	n, _ := strconv.ParseInt(f.StatusID, 10, 64)
	return n
}

// QuantityInt64 returns the parsed Quantity as int64. Returns 1 when blank or invalid.
func (f *UnitForm) QuantityInt64() int64 {
	n, _ := strconv.ParseInt(f.Quantity, 10, 64)
	if n <= 0 {
		return 1
	}
	return n
}

// PurchasePriceCents parses a decimal price string (e.g. "19.99") into integer cents.
// Returns nil when blank or unparseable.
func (f *UnitForm) PurchasePriceCents() *int64 {
	s := strings.ReplaceAll(f.PurchasePrice, ",", ".")
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	cents := int64(math.Round(v * 100))
	return &cents
}

// PurchasedAtUnix parses the YYYY-MM-DD date and returns a *int64 Unix timestamp, or nil when blank.
func (f *UnitForm) PurchasedAtUnix() *int64 {
	if f.PurchasedAt == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", f.PurchasedAt)
	if err != nil {
		return nil
	}
	v := t.UTC().Unix()
	return &v
}

// NextInspectionAtUnix parses the YYYY-MM-DD date and returns a *int64 Unix timestamp, or nil when blank.
func (f *UnitForm) NextInspectionAtUnix() *int64 {
	if f.NextInspectionAt == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", f.NextInspectionAt)
	if err != nil {
		return nil
	}
	v := t.UTC().Unix()
	return &v
}

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
	Count            string // total_stock for bulk; number of units to generate for serialized
	HasContent       string // "1" when checked, "" when unchecked
	PurchasePrice    string
	RentalPrice      string
	Notes            string
	WeightG          string
	WidthMM          string
	HeightMM         string
	DepthMM          string
	PowerW           string
	CurrentA         string
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
		Count:            strings.TrimSpace(r.PostForm.Get("count")),
		HasContent:       r.PostForm.Get("has_content"),
		PurchasePrice:    strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:      strings.TrimSpace(r.PostForm.Get("rental_price")),
		Notes:            strings.TrimSpace(r.PostForm.Get("notes")),
		WeightG:          strings.TrimSpace(r.PostForm.Get("weight_g")),
		WidthMM:          strings.TrimSpace(r.PostForm.Get("width_mm")),
		HeightMM:         strings.TrimSpace(r.PostForm.Get("height_mm")),
		DepthMM:          strings.TrimSpace(r.PostForm.Get("depth_mm")),
		PowerW:           strings.TrimSpace(r.PostForm.Get("power_w")),
		CurrentA:         strings.TrimSpace(r.PostForm.Get("current_a")),
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
	if validator.NotBlank(f.Count) {
		n, err := strconv.ParseInt(f.Count, 10, 64)
		f.Check(err == nil, "count", "Must be a whole number")
		f.Check(err != nil || n >= 1, "count", "Must be at least 1")
	} else {
		f.AddError("count", "This field cannot be blank")
	}

	if validator.NotBlank(f.PurchasePrice) {
		_, err := strconv.ParseFloat(normalizeDecimal(f.PurchasePrice), 64)
		f.Check(err == nil, "purchase_price", "Must be a valid number")
	}

	if validator.NotBlank(f.RentalPrice) {
		_, err := strconv.ParseFloat(normalizeDecimal(f.RentalPrice), 64)
		f.Check(err == nil, "rental_price", "Must be a valid number")
	}

	return f.Valid()
}

// HasContentBool returns true when the has_content checkbox was checked.
func (f *NewForm) HasContentBool() bool { return f.HasContent == "1" }

// WeightGInt64 returns weight_g as *int64 (nil when blank).
func (f *NewForm) WeightGInt64() *int64 { return parseOptionalInt64(f.WeightG) }

// WidthMMInt64 returns width_mm as *int64 (nil when blank).
func (f *NewForm) WidthMMInt64() *int64 { return parseOptionalInt64(f.WidthMM) }

// HeightMMInt64 returns height_mm as *int64 (nil when blank).
func (f *NewForm) HeightMMInt64() *int64 { return parseOptionalInt64(f.HeightMM) }

// DepthMMInt64 returns depth_mm as *int64 (nil when blank).
func (f *NewForm) DepthMMInt64() *int64 { return parseOptionalInt64(f.DepthMM) }

// PowerMW converts the user-entered watts string to milliwatts, or nil when blank.
func (f *NewForm) PowerMW() *int64 { return parseOptionalWattsToMW(f.PowerW) }

// CurrentMA converts the user-entered amps string to milliamps, or nil when blank.
func (f *NewForm) CurrentMA() *int64 { return parseOptionalAmpsToMA(f.CurrentA) }

// CountInt64 returns the parsed Count value. Call only after Validate() returns true.
func (f *NewForm) CountInt64() int64 {
	n, _ := strconv.ParseInt(f.Count, 10, 64)
	return n
}

// PurchasePriceCents returns the purchase price in the smallest currency unit, or nil when blank.
func (f *NewForm) PurchasePriceCents() *int64 {
	return parseOptionalCents(f.PurchasePrice)
}

// RentalPriceCents returns the rental price in the smallest currency unit, or nil when blank.
func (f *NewForm) RentalPriceCents() *int64 {
	return parseOptionalCents(f.RentalPrice)
}

// parseOptionalCents parses a decimal price string (accepting both "." and "," as
// the decimal separator) and returns the value in the smallest currency unit (cents).
// Returns nil when s is blank or unparseable.
func parseOptionalCents(s string) *int64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(normalizeDecimal(s), 64)
	if err != nil {
		return nil
	}
	v := int64(math.Round(f * 100))
	return &v
}

// parseOptionalInt64 parses a whole-number string. Returns nil when s is blank or unparseable.
func parseOptionalInt64(s string) *int64 {
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

// parseOptionalWattsToMW parses a decimal watts string and returns milliwatts. Returns nil when blank.
func parseOptionalWattsToMW(s string) *int64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(normalizeDecimal(s), 64)
	if err != nil {
		return nil
	}
	v := int64(math.Round(f * 1000))
	return &v
}

// ContentForm holds parsed input and validation state for assigning content to an equipment item.
type ContentForm struct {
	MemberName string
	Quantity   string
	validator.Validator
}

// ParseContent reads the content assign form fields from r.
func ParseContent(r *http.Request) (ContentForm, error) {
	if err := r.ParseForm(); err != nil {
		return ContentForm{}, fmt.Errorf("parse form: %w", err)
	}
	return ContentForm{
		MemberName: strings.TrimSpace(r.PostForm.Get("member_name")),
		Quantity:   strings.TrimSpace(r.PostForm.Get("quantity")),
	}, nil
}

// Validate checks ContentForm fields and returns true when all checks pass.
func (f *ContentForm) Validate() bool {
	f.Check(validator.NotBlank(f.MemberName), "member_name", "An item must be selected")
	if validator.NotBlank(f.Quantity) {
		n, err := strconv.ParseInt(f.Quantity, 10, 64)
		f.Check(err == nil, "quantity", "Must be a whole number")
		f.Check(err != nil || n >= 1, "quantity", "Must be at least 1")
	} else {
		f.AddError("quantity", "This field cannot be blank")
	}
	return f.Valid()
}

// QuantityInt64 returns the parsed Quantity. Call only after Validate() returns true.
func (f *ContentForm) QuantityInt64() int64 {
	n, _ := strconv.ParseInt(f.Quantity, 10, 64)
	return n
}

// parseOptionalAmpsToMA parses a decimal amps string and returns milliamps. Returns nil when blank.
func parseOptionalAmpsToMA(s string) *int64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(normalizeDecimal(s), 64)
	if err != nil {
		return nil
	}
	v := int64(math.Round(f * 1000))
	return &v
}
