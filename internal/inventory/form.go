package inventory

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

const maxUploadBytes = 10 << 20 // 10 MiB

// Form holds the parsed form input and validation state for inventory create/update requests.
type Form struct {
	Name        string
	CategoryID  string
	TotalStock  string
	Notes       string
	Image       multipart.File // nil when no file was uploaded
	ImageHeader *multipart.FileHeader
	validator.Validator
}

// Parse reads the inventory form fields from r, including an optional image upload.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil { //nolint:gosec // maxUploadBytes is a bounded constant (10 MiB)
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	f := Form{
		Name:       strings.TrimSpace(r.PostForm.Get("name")),
		CategoryID: strings.TrimSpace(r.PostForm.Get("category_id")),
		TotalStock: strings.TrimSpace(r.PostForm.Get("total_stock")),
		Notes:      strings.TrimSpace(r.PostForm.Get("notes")),
	}
	file, header, err := r.FormFile("image")
	if err == nil {
		f.Image = file
		f.ImageHeader = header
	}
	return f, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
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

	return f.Valid()
}

// TotalStockInt64 returns the parsed TotalStock value. Call only after Validate() returns true.
func (f *Form) TotalStockInt64() int64 {
	n, _ := strconv.ParseInt(f.TotalStock, 10, 64)
	return n
}
