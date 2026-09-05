// Copyright (C) 2026 bit8bytes
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
	"net/url"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/locale"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
	"golang.org/x/text/language"
)

func (app *application) getLanding(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	if locale.TagFrom(r.Context()) == language.German {
		return app.html.Render(w, r, http.StatusOK, pages.LandingDE, data)
	}
	return app.html.Render(w, r, http.StatusOK, pages.Landing, data)
}

func (app *application) postLocale(w http.ResponseWriter, r *http.Request) {
	locale := r.FormValue("locale")
	http.SetCookie(w, &http.Cookie{
		Name:     localeCookieName,
		Value:    locale,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	ref := "/"
	if u, err := url.Parse(r.Referer()); err == nil && u.Path != "" {
		safe := url.URL{Path: u.Path, RawQuery: u.RawQuery}
		ref = safe.String()
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}

func (app *application) getImprint(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	return app.html.Render(w, r, http.StatusOK, pages.Imprint, data)
}

func (app *application) getPrivacy(w http.ResponseWriter, r *http.Request) *httperr.Error {
	data := app.html.TemplateData(r)
	return app.html.Render(w, r, http.StatusOK, pages.Privacy, data)
}

func getLLMsTxt(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = fmt.Fprint(w, "# Gearberg\n\nOpen-source equipment tracking and rental management. Self-host with Docker. Licensed under AGPL-3.0.\n\n## Links\n\n- [GitHub](https://github.com/bit8bytes/gearberg)\n- [Specification](https://github.com/bit8bytes/gearberg/blob/main/wiki/SPECS.md)\n- [License](https://www.gnu.org/licenses/agpl-3.0.html)\n")
}

func getRobotsTxt(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = fmt.Fprint(w, "User-agent: *\nAllow: /\n")
}
