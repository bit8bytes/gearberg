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

// Package html provides the Renderer type, which adapts httperr.HandlerFunc into standard
// http.HandlerFunc and renders HTML page templates with buffered writes.
package html

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/nonce"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"github.com/bit8bytes/gearberg/internal/trace"
)

// TemplateData is the base data passed to every page template.
type TemplateData struct {
	// Form holds data for interactive HTML forms (e.g. POST/GET on new and edit pages).
	// Use Form when the page has user input that can change, such as validation errors or prefilled values.
	Form any
	// Data holds read-only data passed to the page for display.
	// Unlike Form, Data does not change based on user interaction.
	Data any
	// URL passes the current request URL with query parameters to the template.
	// It is used for highlighting the active navigation path.
	URL *url.URL
	// Nonce is the per-request CSP nonce injected into script and style tags.
	Nonce nonce.Nonce
	// Revision is the binary revision string used for cache-busting static asset URLs.
	Revision string
	// OrgID is the active organisation from the URL path, populated on org-scoped pages.
	// Empty on user-level pages (e.g. /settings/account). Used by the header partial.
	OrgID string
}

// HTML renders HTML responses and adapts httperr.HandlerFunc into standard http.HandlerFunc.
type HTML struct {
	logger   *slog.Logger
	base     *template.Template
	cache    map[string]*template.Template
	revision string
}

// New returns a Renderer backed by the given logger, template cache, and revision string.
func New(logger *slog.Logger, base *template.Template, cache map[string]*template.Template, revision string) *HTML {
	return &HTML{logger: logger, base: base, cache: cache, revision: revision}
}

// Handle adapts an httperr.HandlerFunc into a standard http.HandlerFunc. When the handler
// returns a non-nil *httperr.Error, it logs the error and renders the error page.
func (rnd *HTML) Handle(fn httperr.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			traceID := trace.From(r.Context())
			err.TraceID = traceID.String()
			rnd.logger.ErrorContext(r.Context(), "handler error",
				slog.String("message", err.Message),
				slog.Any("err", err.Error),
				slog.String("trace_id", err.TraceID),
			)
			errData := rnd.TemplateData(r)
			errData.Data = err
			if renderErr := rnd.Render(w, r, err.Code, pages.Error, errData); renderErr != nil {
				rnd.logger.ErrorContext(r.Context(), "render error page", slog.Any("err", renderErr.Error))
				http.Error(w, err.Message, err.Code)
			}
		}
	})
}

// Render executes the page template into a buffer and writes it to w with the given status code.
// Buffering prevents a partial write if template execution fails after headers are sent.
func (rnd *HTML) Render(w http.ResponseWriter, _ *http.Request, status int, page pages.Page, data any) *httperr.Error {
	tmpl, ok := rnd.cache[page.File]
	if !ok {
		return httperr.InternalServerError(fmt.Errorf("template not found: %v", page.File))
	}

	buffer := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(buffer, "root", data); err != nil {
		return httperr.InternalServerError(fmt.Errorf("render: %w", err))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if _, err := buffer.WriteTo(w); err != nil {
		return httperr.InternalServerError(fmt.Errorf("render write: %w", err))
	}
	return nil
}

// RenderFragment executes a named fragment block directly from the base template,
// returning only the fragment HTML with no surrounding page shell.
func (rnd *HTML) RenderFragment(w http.ResponseWriter, _ *http.Request, status int, name string, data any) *httperr.Error {
	buffer := new(bytes.Buffer)
	if err := rnd.base.ExecuteTemplate(buffer, name, data); err != nil {
		return httperr.InternalServerError(fmt.Errorf("render fragment %q: %w", name, err))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	if _, err := buffer.WriteTo(w); err != nil {
		return httperr.InternalServerError(fmt.Errorf("render fragment write %q: %w", name, err))
	}
	return nil
}

// TemplateData builds the base template data shared by every page.
func (rnd *HTML) TemplateData(r *http.Request) *TemplateData {
	return &TemplateData{
		URL:      r.URL,
		Nonce:    nonce.From(r.Context()),
		Revision: rnd.revision,
		OrgID:    r.PathValue("org_id"),
	}
}
