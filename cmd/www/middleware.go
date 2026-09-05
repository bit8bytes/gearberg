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
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/bit8bytes/gearberg/internal/locale"
	"github.com/bit8bytes/gearberg/internal/nonce"
	"github.com/bit8bytes/gearberg/internal/trace"
	"github.com/bit8bytes/gearberg/pkg/tokens"
	"golang.org/x/text/language"
)

func withLocale(next http.Handler) http.Handler {
	tags := []language.Tag{language.English, language.German}
	matcher := language.NewMatcher(tags)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := "en-US"
		if c, err := r.Cookie(localeCookieName); err == nil {
			loc = c.Value
		}
		tag, _ := language.MatchStrings(matcher, loc, r.Header.Get("Accept-Language"))
		ctx := context.WithValue(r.Context(), locale.Key{}, tag)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := trace.NewContext(r.Context(), tokens.Generate().Hex())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withNonce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := nonce.NewContext(r.Context(), tokens.Generate().Hex())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type requestLogger struct {
	logger *slog.Logger
}

func newRequestLogger(logger *slog.Logger) *requestLogger {
	return &requestLogger{logger: logger}
}

// logRequest is middleware that logs each inbound request. Requests to /s/ (static assets)
// are skipped to avoid flooding logs with high-frequency, low-signal noise.
func (rl *requestLogger) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		rl.logger.InfoContext(r.Context(), "request",
			slog.String("addr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("queries", r.URL.Query()),
		)

		next.ServeHTTP(w, r)
	})
}

type panicRecoverer struct {
	logger *slog.Logger
}

func newPanicRecoverer(logger *slog.Logger) *panicRecoverer {
	return &panicRecoverer{logger: logger}
}

// recoverPanic catches any panics that occur in downstream handlers, closes the
// connection to prevent the client from hanging on a half-written response.
// Without this, an unrecovered panic would crash the goroutine and leave the request
// silently unanswered.
func (pr *panicRecoverer) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				stack := debug.Stack()

				var panicErr error
				switch v := err.(type) {
				case error:
					panicErr = v
				case string:
					panicErr = errors.New(v)
				default:
					panicErr = fmt.Errorf("panic: %v", err)
				}

				traceID := trace.From(r.Context())

				pr.logger.Error("recover from panic",
					slog.String("type", "panic"),
					slog.String("error", panicErr.Error()),
					slog.String("type", fmt.Sprintf("%T", err)),
					slog.String("stack", string(stack)),
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
					slog.String("trace_id", traceID.String()),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// withMaxBodySize caps the request body at 32 MiB to prevent memory exhaustion
// from oversized form submissions or request bodies.
func withMaxBodySize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
		next.ServeHTTP(w, r)
	})
}

// withSecurityHeaders sets HTTP response headers that harden the app against
// common browser-based attacks. Must be applied to all routes.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000;")

		nonceID := nonce.From(r.Context())
		w.Header().Set("Content-Security-Policy", fmt.Sprintf(
			"default-src 'self'; "+
				"img-src 'self' data:;"+
				"style-src 'nonce-%s'; "+
				"script-src 'nonce-%s' 'strict-dynamic' https://gc.zgo.at; "+
				"font-src 'self'; "+
				"connect-src 'self' https://bit8bytes.goatcounter.com; "+
				"form-action 'self'; "+
				"frame-ancestors 'none'; "+
				"object-src 'none'; "+
				"base-uri 'self';",
			nonceID, nonceID.String(),
		))

		next.ServeHTTP(w, r)
	})
}
