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
	"fmt"
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/internal/assets"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = fmt.Fprint(w, "User-agent: *\nAllow: /\n")
	})

	mux.HandleFunc("/", app.html.Handle(app.getDocs))
	mux.Handle("GET /imprint", http.RedirectHandler("https://www.gearberg.org/imprint", http.StatusSeeOther))
	mux.Handle("GET /privacy", http.RedirectHandler("https://www.gearberg.org/privacy", http.StatusSeeOther))

	logRequest := newRequestLogger(app.logger)
	recoverPanic := newPanicRecoverer(app.logger)

	return withTrace(
		withNonce(
			recoverPanic.handler(
				logRequest.handler(
					withSecurityHeaders(mux)))))
}

func (app *application) getDocs(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Data = struct {
		Year int
	}{
		Year: time.Now().Year(),
	}
	return app.html.Render(w, r, http.StatusOK, pages.Docs, data)
}
