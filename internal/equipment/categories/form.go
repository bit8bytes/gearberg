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

// Package categories handles equipment category routes, business logic, and data access.
package categories

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bit8bytes/toolbox/validator"
)

// Form holds the parsed form input and validation state for category create/update requests.
type Form struct {
	Name string
	validator.Validator
}

// NewForm returns a Form with an initialized Errors map, safe for template rendering.
func NewForm() *Form {
	f := &Form{}
	f.Errors = make(map[string]string)
	return f
}

// Parse reads the category form fields from r.
func Parse(r *http.Request) (Form, error) {
	f := Form{}
	f.Errors = make(map[string]string)
	if err := r.ParseForm(); err != nil {
		return f, fmt.Errorf("parse form: %w", err)
	}
	f.Name = strings.TrimSpace(r.PostForm.Get("name"))
	return f, nil
}

// Validate checks form fields and returns true when all checks pass.
func (f *Form) Validate() bool {
	f.Check(validator.NotBlank(f.Name), "name", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Name, 100), "name", "This field cannot exceed 100 characters")
	return f.Valid()
}
