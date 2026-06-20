package locations

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

// Form holds the parsed form input and validation state for location create/update requests.
type Form struct {
	Name string
	validator.Validator
}

// Parse reads the location form fields from r.
func Parse(r *http.Request) (Form, error) {
	if err := r.ParseForm(); err != nil {
		return Form{}, fmt.Errorf("parse form: %w", err)
	}
	return Form{
		Name: strings.TrimSpace(r.PostForm.Get("name")),
	}, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 100), "name", "This field cannot exceed 100 characters")
	return f.Valid()
}
