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
	"net/http"

	"github.com/bit8bytes/gearberg/internal/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))
	mux.HandleFunc("GET /robots.txt", getRobotsTxt)
	mux.HandleFunc("GET /llms.txt", getLLMsTxt)

	mux.HandleFunc("/", app.html.Handle(app.getLanding))
	mux.HandleFunc("GET /imprint", app.html.Handle(app.getImprint))
	mux.HandleFunc("GET /privacy", app.html.Handle(app.getPrivacy))
	mux.HandleFunc("POST /locale", app.postLocale)

	antiCSRF := http.NewCrossOriginProtection()
	logRequest := newRequestLogger(app.logger)
	recoverPanic := newPanicRecoverer(app.logger)

	return withTrace(
		withNonce(
			recoverPanic.handler(
				logRequest.handler(
					withSecurityHeaders(
						withMaxBodySize(
							antiCSRF.Handler(
								withLocale(mux))))))))
}
