// Package httperr defines the error type and handler func signature for HTTP handlers
// that return structured errors instead of writing error responses directly.
package httperr

import "net/http"

// Error carries an internal error, a user-facing message, and an HTTP status code together
// so handlers can return a single value that encodes all three.
type Error struct {
	Error   error
	Message string
	Code    int
	TraceID string
}

// HandlerFunc is an http.HandlerFunc variant that returns an *Error instead of writing error
// responses directly. Handlers return nil on success; a non-nil value is handled by the renderer.
type HandlerFunc func(http.ResponseWriter, *http.Request) *Error
