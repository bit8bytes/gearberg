package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/bit8bytes/gearberg/templates/pages"
)

// appError carries an internal error, a user-facing message, and an HTTP status code together
// so handlers can return a single value that encodes all three.
type appError struct {
	Error   error
	Message string
	Code    int
}

// appHandler is an http.HandlerFunc variant that returns an *appError instead of writing error
// responses directly. Handlers return nil on success; a non-nil value is handled by handleHTML.
type appHandler func(http.ResponseWriter, *http.Request) *appError

// handleHTML adapts an appHandler into a standard http.HandlerFunc. If the handler returns a
// non-nil *appError, it renders the error page with the status code and message from that error.
func (app *application) handleHTML(fn appHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			app.logger.ErrorContext(r.Context(), "handler error",
				slog.Int("status", err.Code),
				slog.String("message", err.Message),
				slog.Any("err", err.Error),
			)
			if renderErr := app.render(w, r, err.Code, pages.Error, err); renderErr != nil {
				app.logger.ErrorContext(r.Context(), "render error page", slog.Any("err", renderErr.Error))
				http.Error(w, err.Message, err.Code)
			}
		}
	})
}

// render renders the given page into a buffer and writes it to w with status. It returns an
// *appError so it can be used directly inside appHandler functions without an adapter.
func (app *application) render(w http.ResponseWriter, _ *http.Request, status int, page pages.Page, data any) *appError {
	tmpl, ok := app.cache[page.File]
	if !ok {
		return &appError{
			Error:   fmt.Errorf("template not found: %v", page.File),
			Message: "Template not found.",
			Code:    http.StatusNotFound,
		}
	}

	buffer, err := renderToBuffer(tmpl, "root", data)
	if err != nil {
		return &appError{
			Error:   fmt.Errorf("render: %w", err),
			Message: "Failed to render page.",
			Code:    http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	_, err = buffer.WriteTo(w)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to write response.",
			Code:    http.StatusInternalServerError,
		}
	}
	return nil
}

// renderToBuffer executes the named template into an in-memory buffer instead of writing
// directly to the ResponseWriter. This allows the caller to inspect or discard the output
// before committing to a response — once WriteHeader is called the status code cannot change.
func renderToBuffer(tmpl *template.Template, name string, data any) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(buffer, name, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buffer, nil
}

type templateData struct {
	// Form holds data for interactive HTML forms (e.g. POST/GET on new and edit pages).
	// Use Form when the page has user input that can change, such as validation errors or prefilled values.
	Form any
	// Data holds read-only data passed to the page for display.
	// Unlike Form, Data does not change based on user interaction.
	Data any
	// URL passes the current request URL with query parameters to the template.
	// It is used for highlighting the active navigation path.
	URL *url.URL
}

// newTemplateData builds the base template data shared by every page: the request URL
// (for active-nav highlighting), the CSP nonce (injected into inline scripts/styles),
// and the binary revision (for cache-busting static assets).
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		URL: r.URL,
	}
}
