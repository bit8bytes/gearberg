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

// BadRequest returns a 400 error with a generic "Bad request." message.
func BadRequest(err error) *Error {
	return &Error{
		Error:   err,
		Message: "Bad request.",
		Code:    http.StatusBadRequest,
	}
}

// InternalServerError returns a 500 error with a generic "Something went wrong." message.
// Do not expose detail in the message. Log it via the Error field instead.
func InternalServerError(err error) *Error {
	return &Error{
		Error:   err,
		Message: "Something went wrong.",
		Code:    http.StatusInternalServerError,
	}
}

// NotFound returns a 404 error with a generic  "Not found." message.
// Do not expose deail in message. Log it via the Error field instead.
func NotFound(err error) *Error {
	return &Error{
		Error:   err,
		Message: "Not found.",
		Code:    http.StatusNotFound,
	}
}
