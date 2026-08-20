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

// Package templates holds the HTML templates embedded into the binary so the
// application can be deployed as a single self-contained executable without
// requiring template files to exist on the host filesystem.
package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"

	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

// EmbedFS contains all template directories embedded at compile time.
//
//go:embed components layouts pages partials fragments
var EmbedFS embed.FS

// Templates holds the funcMap shared across all parsed template sets.
type Templates struct {
	funcMap template.FuncMap
}

// Option configures a Templates at construction time.
type Option func(*Templates)

// WithFuncs injects custom template functions before parsing.
func WithFuncs(funcs template.FuncMap) Option {
	return func(t *Templates) {
		t.funcMap = funcs
	}
}

// New initializes a Templates with an empty funcMap so callers never need to
// nil-check before adding functions.
func New(opts ...Option) *Templates {
	t := &Templates{funcMap: template.FuncMap{}}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Parse returns the shared base (for fragment rendering) and a per-page map
// (for full-page rendering) so both render paths share one compiled template set.
func (t *Templates) Parse(pages []pages.Page) (*template.Template, map[string]*template.Template, error) {
	// missingkey=error surfaces missing map keys at render time instead of silently emitting "".
	base, err := template.New("root").Funcs(t.funcMap).Option("missingkey=error").ParseFS(EmbedFS, "layouts/root.tmpl", "components/*.tmpl", "fragments/*.tmpl")
	if err != nil {
		return nil, nil, fmt.Errorf("base template: %w", err)
	}

	tmpls := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		t, err := pageTemplate(EmbedFS, base, page)
		if err != nil {
			return nil, nil, fmt.Errorf("page template %s: %w", page.File, err)
		}
		tmpls[page.File] = t
	}
	return base, tmpls, nil
}

func pageTemplate(fsys fs.FS, base *template.Template, page pages.Page) (*template.Template, error) {
	// Clone so adding page-specific templates doesn't mutate the shared base,
	// which would cause data races under concurrent requests.
	t := template.Must(base.Clone())

	patterns := []string{page.Layout.File, page.File}
	if page.Layout.Partials != "" {
		// Empty pattern strings cause ParseFS to error, so only append when set.
		patterns = append(patterns, page.Layout.Partials)
	}

	t, err := t.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	// Missing "page" block renders silently empty inside the layout.
	if t.Lookup("page") == nil {
		return nil, fmt.Errorf("page block not defined: %s", page.File)
	}

	return t, nil
}
