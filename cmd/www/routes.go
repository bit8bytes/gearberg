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
	mux.HandleFunc("GET /llms.txt", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = fmt.Fprint(w, "# Gearberg\n\nOpen-source equipment tracking and rental management. Self-host with Docker. Licensed under AGPL-3.0.\n\n## Links\n\n- [GitHub](https://github.com/bit8bytes/gearberg)\n- [Specification](https://github.com/bit8bytes/gearberg/blob/main/wiki/SPECS.md)\n- [License](https://www.gnu.org/licenses/agpl-3.0.html)\n")
	})

	mux.HandleFunc("/", app.html.Handle(app.getLanding))
	mux.HandleFunc("GET /imprint", app.html.Handle(app.getImprint))
	mux.HandleFunc("GET /privacy", app.html.Handle(app.getPrivacy))

	antiCSRF := http.NewCrossOriginProtection()
	logRequest := newRequestLogger(app.logger)
	recoverPanic := newPanicRecoverer(app.logger)

	return withTrace(
		withNonce(
			recoverPanic.handler(
				logRequest.handler(
					withSecurityHeaders(
						withMaxBodySize(
							antiCSRF.Handler(mux)))))))
}

func (app *application) getLanding(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Data = struct {
		Year int
	}{
		Year: time.Now().Year(),
	}
	return app.html.Render(w, r, http.StatusOK, pages.Landing, data)
}

func (app *application) getImprint(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Data = struct{ Year int }{Year: time.Now().Year()}
	return app.html.Render(w, r, http.StatusOK, pages.Imprint, data)
}

func (app *application) getPrivacy(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	data.Data = struct{ Year int }{Year: time.Now().Year()}
	return app.html.Render(w, r, http.StatusOK, pages.Privacy, data)
}
