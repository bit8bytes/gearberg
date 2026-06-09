package inventory

import (
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

const maxUploadBytes = 10 << 20 // 10 MiB

// Form holds the parsed form input and validation state for inventory update requests.
type Form struct {
	Name           string
	CategoryID     string
	ManufacturerID string
	Code           string
	TotalStock     string
	PurchasePrice  string
	RentalPrice    string
	Notes          string
	WeightG        string
	WidthMM        string
	HeightMM       string
	DepthMM        string
	PowerW         string // user enters watts; stored as milliwatts
	CurrentA       string // user enters amps; stored as milliamps
	Image          multipart.File
	ImageHeader    *multipart.FileHeader
	validator.Validator
}

// normalizeDecimal replaces comma with dot so both European and US number
// formats are accepted before parsing (e.g. "19,99" → "19.99").
func normalizeDecimal(s string) string {
	return strings.ReplaceAll(s, ",", ".")
}

// Parse reads the inventory form fields from r, including an optional image upload.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	f := Form{
		Name:           strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID:     strings.TrimSpace(r.PostForm.Get("category_id")),
		ManufacturerID: strings.TrimSpace(r.PostForm.Get("manufacturer_id")),
		Code:           strings.TrimSpace(r.PostForm.Get("code")),
		TotalStock:     strings.TrimSpace(r.PostForm.Get("total_stock")),
		PurchasePrice:  strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:    strings.TrimSpace(r.PostForm.Get("rental_price")),
		Notes:          strings.TrimSpace(r.PostForm.Get("notes")),
		WeightG:        strings.TrimSpace(r.PostForm.Get("weight_g")),
		WidthMM:        strings.TrimSpace(r.PostForm.Get("width_mm")),
		HeightMM:       strings.TrimSpace(r.PostForm.Get("height_mm")),
		DepthMM:        strings.TrimSpace(r.PostForm.Get("depth_mm")),
		PowerW:         strings.TrimSpace(r.PostForm.Get("power_w")),
		CurrentA:       strings.TrimSpace(r.PostForm.Get("current_a")),
	}
	file, header, err := r.FormFile("image")
	if err == nil {
		f.Image = file
		f.ImageHeader = header
	}
	return f, nil
}

// CodeInt64 returns the parsed Code value as int64. Call only after ValidateBulk or ValidateSerialized returns true.
func (f *Form) CodeInt64() int64 {
	n, _ := strconv.ParseInt(f.Code, 10, 64)
	return n
}

func (f *Form) validateCode() {
	if validator.NotBlank(f.Code) {
		n, err := strconv.ParseInt(f.Code, 10, 64)
		f.Check(err == nil, "code", "Must be a whole number")
		f.Check(err != nil || n >= 1, "code", "Must be at least 1")
	} else {
		f.AddError("code", "This field cannot be blank")
	}
}

// ValidateBulk checks form fields for a bulk inventory update and returns true when all pass.
func (f *Form) ValidateBulk() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")
	f.Check(validator.NotBlank(f.CategoryID), "category_id", "A category must be selected")
	f.validateCode()

	if validator.NotBlank(f.TotalStock) {
		n, err := strconv.ParseInt(f.TotalStock, 10, 64)
		f.Check(err == nil, "total_stock", "Must be a whole number")
		f.Check(err != nil || n >= 1, "total_stock", "Must be at least 1")
	} else {
		f.AddError("total_stock", "This field cannot be blank")
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

// ValidateSerialized checks form fields for a serialized inventory update and returns true when all pass.
// total_stock is not validated because it is derived from the unit count.
func (f *Form) ValidateSerialized() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")
	f.Check(validator.NotBlank(f.CategoryID), "category_id", "A category must be selected")
	f.validateCode()

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

// TotalStockInt64 returns the parsed TotalStock value. Call only after ValidateBulk() returns true.
func (f *Form) TotalStockInt64() int64 {
	n, _ := strconv.ParseInt(f.TotalStock, 10, 64)
	return n
}

// PurchasePriceCents returns the purchase price in the smallest currency unit, or nil when blank.
func (f *Form) PurchasePriceCents() *int64 {
	return parseOptionalCents(f.PurchasePrice)
}

// RentalPriceCents returns the rental price in the smallest currency unit, or nil when blank.
func (f *Form) RentalPriceCents() *int64 {
	return parseOptionalCents(f.RentalPrice)
}

// WeightGInt64 returns weight_g as *int64 (nil when blank).
func (f *Form) WeightGInt64() *int64 { return parseOptionalInt64(f.WeightG) }

// WidthMMInt64 returns width_mm as *int64 (nil when blank).
func (f *Form) WidthMMInt64() *int64 { return parseOptionalInt64(f.WidthMM) }

// HeightMMInt64 returns height_mm as *int64 (nil when blank).
func (f *Form) HeightMMInt64() *int64 { return parseOptionalInt64(f.HeightMM) }

// DepthMMInt64 returns depth_mm as *int64 (nil when blank).
func (f *Form) DepthMMInt64() *int64 { return parseOptionalInt64(f.DepthMM) }

// PowerMW converts the user-entered watts string to milliwatts, or nil when blank.
func (f *Form) PowerMW() *int64 { return parseOptionalWattsToMW(f.PowerW) }

// CurrentMA converts the user-entered amps string to milliamps, or nil when blank.
func (f *Form) CurrentMA() *int64 { return parseOptionalAmpsToMA(f.CurrentA) }

// NewForm holds the parsed form input and validation state for inventory creation (both types).
type NewForm struct {
	TypeID           string // "bulk" or "serialized"
	UsageTypeID      string // "rental" or "sale"
	Name             string
	CategoryID       string
	CategoryName     string // set when user typed a new category name not yet in the DB
	ManufacturerID   string
	ManufacturerName string // set when user typed a new manufacturer name not yet in the DB
	Count            string // total_stock for bulk; number of units to generate for serialized
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
		Name:             strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID:       strings.TrimSpace(r.PostForm.Get("category_id")),
		CategoryName:     strings.TrimSpace(r.PostForm.Get("category_name")),
		ManufacturerID:   strings.TrimSpace(r.PostForm.Get("manufacturer_id")),
		ManufacturerName: strings.TrimSpace(r.PostForm.Get("manufacturer_name")),
		Count:            strings.TrimSpace(r.PostForm.Get("count")),
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
	f.Check(validator.NotBlank(f.CategoryID) || validator.NotBlank(f.CategoryName), "category_id", "A category must be selected")

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
