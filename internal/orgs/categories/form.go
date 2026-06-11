// Package categories handles equipment category routes, business logic, and data access.
package categories

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

var hexColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Form holds the parsed form input and validation state for category create/update requests.
type Form struct {
	Name  string
	Color string
	validator.Validator
}

// Parse reads the category form fields from r.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseForm(); err != nil {
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	color := r.PostForm.Get("color")
	if color == "" {
		color = DefaultColor
	}
	return Form{
		Name:  strings.TrimSpace(r.PostForm.Get("name")),
		Color: color,
	}, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 100), "name", "This field cannot exceed 100 characters")
	f.Check(hexColorRE.MatchString(f.Color), "color", "Must be a valid hex color (e.g. #FF5733)")
	return f.Valid()
}
