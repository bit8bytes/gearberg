// Package json provides helpers for reading and writing JSON over HTTP.
package json

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"strings"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/trace"
)

// JSON wraps common JSON request/response operations for HTTP handlers.
type JSON struct {
	logger *slog.Logger
}

// New returns a JSON helper that logs errors via logger.
func New(logger *slog.Logger) *JSON {
	return &JSON{logger: logger}
}

// Handle wraps fn so that any returned error is logged and written as a JSON error response.
func (jsn *JSON) Handle(fn httperr.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			traceID := trace.From(r.Context())
			err.TraceID = traceID.String()
			jsn.logger.ErrorContext(r.Context(), "handler error",
				slog.Int("status", err.Code),
				slog.String("message", err.Message),
				slog.Any("err", err.Error),
				slog.String("trace_id", err.TraceID),
			)
			if writeErr := jsn.Write(r.Context(), w, err.Code, map[string]string{"error": err.Message}, nil); writeErr != nil {
				jsn.logger.ErrorContext(r.Context(), "write error response", slog.Any("err", writeErr.Error))
				http.Error(w, err.Message, err.Code)
			}
		}
	})
}

// Write marshals data as JSON and writes it to w with the given status code and headers.
func (jsn *JSON) Write(ctx context.Context, w http.ResponseWriter, status int, data any, headers http.Header) *httperr.Error {
	js, err := json.Marshal(data)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "failed to encode JSON response",
			Code:    http.StatusInternalServerError,
		}
	}

	traceID := trace.From(ctx)
	w.Header().Set("X-Trace-Id", traceID.String())
	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(js)

	return nil
}

// DefaultMaxBytes is the default request body size limit used with Parse (1 MiB).
const DefaultMaxBytes int64 = 1_048_576

// Parse decodes a single JSON value from r.Body into dst.
// maxBytes caps the request body size; use DefaultMaxBytes for the standard 1 MiB limit.
func Parse(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return translateDecodeError(err)
	}

	err := dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// translateDecodeError converts raw json.Decoder errors into user-facing messages.
// Decode() returns "json: unknown field <name>" for unknown fields; see
// https://github.com/golang/go/issues/29035 for a future distinct error type.
func translateDecodeError(err error) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var maxBytesError *http.MaxBytesError

	switch {
	case errors.As(err, &syntaxError):
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
		}
		return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("body contains unknown key %s", fieldName)
	case errors.As(err, &maxBytesError):
		return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
	case errors.As(err, &invalidUnmarshalError):
		panic(err)
	default:
		return fmt.Errorf("Parse: %w", err)
	}
}
