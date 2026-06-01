package inventory

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

// Form holds the parsed form input and validation state for inventory update requests.
type Form struct {
	Name          string
	CategoryID    string
	Code          string
	TotalStock    string
	PurchasePrice string
	RentalPrice   string
	Notes         string
	Image         multipart.File // nil when no file was uploaded
	ImageHeader   *multipart.FileHeader
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
		Name:          strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID:    strings.TrimSpace(r.PostForm.Get("category_id")),
		Code:          strings.TrimSpace(r.PostForm.Get("code")),
		TotalStock:    strings.TrimSpace(r.PostForm.Get("total_stock")),
		PurchasePrice: strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:   strings.TrimSpace(r.PostForm.Get("rental_price")),
		Notes:         strings.TrimSpace(r.PostForm.Get("notes")),
	}
	file, header, err := r.FormFile("image")
	if err == nil {
		f.Image = file
		f.ImageHeader = header
	}
	return f, nil
}

// ValidateBulk checks form fields for a bulk inventory update and returns true when all pass.
func (f *Form) ValidateBulk() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 200), "name", "This field cannot exceed 200 characters")
	f.Check(validator.NotBlank(f.CategoryID), "category_id", "A category must be selected")

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

// NewForm holds the parsed form input and validation state for inventory creation (both types).
type NewForm struct {
	TypeID        string // "bulk" or "serialized"
	UsageTypeID   string // "rental" or "sale"
	Name          string
	CategoryID    string
	Code          string
	Count         string // total_stock for bulk; number of units to generate for serialized
	PurchasePrice string
	RentalPrice   string
	Notes         string
	Image         multipart.File
	ImageHeader   *multipart.FileHeader
	validator.Validator
}

// ParseNew reads the unified create form fields from r.
func ParseNew(r *http.Request) (NewForm, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return NewForm{}, fmt.Errorf("parse form: %w", err)
	}
	f := NewForm{
		TypeID:        strings.TrimSpace(r.PostForm.Get("type_id")),
		UsageTypeID:   strings.TrimSpace(r.PostForm.Get("usage_type_id")),
		Name:          strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID:    strings.TrimSpace(r.PostForm.Get("category_id")),
		Code:          strings.TrimSpace(r.PostForm.Get("code")),
		Count:         strings.TrimSpace(r.PostForm.Get("count")),
		PurchasePrice: strings.TrimSpace(r.PostForm.Get("purchase_price")),
		RentalPrice:   strings.TrimSpace(r.PostForm.Get("rental_price")),
		Notes:         strings.TrimSpace(r.PostForm.Get("notes")),
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
	f.Check(validator.NotBlank(f.CategoryID), "category_id", "A category must be selected")

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

// UnitForm holds the parsed form input and validation state for adding or editing a unit.
type UnitForm struct {
	SerialNumber     string
	NextInspectionAt string
	validator.Validator
}

// ParseUnit reads unit form fields from r.
func ParseUnit(r *http.Request) (UnitForm, error) {
	if err := r.ParseForm(); err != nil {
		return UnitForm{}, fmt.Errorf("parse form: %w", err)
	}
	return UnitForm{
		SerialNumber:     strings.TrimSpace(r.PostForm.Get("serial_number")),
		NextInspectionAt: strings.TrimSpace(r.PostForm.Get("next_inspection_at")),
	}, nil
}

// Validate checks UnitForm fields and returns true when all checks pass.
func (f *UnitForm) Validate() bool {
	if validator.NotBlank(f.SerialNumber) {
		f.Check(validator.MaxChars(f.SerialNumber, 200), "serial_number", "Cannot exceed 200 characters")
	}
	if validator.NotBlank(f.NextInspectionAt) {
		_, err := time.Parse("2006-01-02", f.NextInspectionAt)
		f.Check(err == nil, "next_inspection_at", "Must be a valid date (YYYY-MM-DD)")
	}
	return f.Valid()
}

// NextInspectionAtUnix returns the parsed next inspection date as a Unix timestamp, or nil when blank.
func (f *UnitForm) NextInspectionAtUnix() *int64 {
	return parseDate(f.NextInspectionAt)
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

// parseDate parses a date string in "2006-01-02" format and returns its Unix timestamp,
// or nil when the string is blank or invalid.
func parseDate(s string) *int64 {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	v := t.UTC().Unix()
	return &v
}
