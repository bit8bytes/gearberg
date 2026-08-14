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

// Package layouts defines the available page layouts.
package layouts

// Layout groups a layout template file with the partials glob it requires.
// Partials is a glob pattern; leave empty when the layout needs no partials.
type Layout struct {
	File     string
	Partials string
}

// Landing renders the public landing page with a simple full-bleed layout.
var Landing = Layout{File: "layouts/landing.tmpl", Partials: "partials/*.tmpl"}

// Docs renders the public docs page with a simple full-bleed layout.
var Docs = Layout{File: "layouts/docs.tmpl", Partials: "partials/docs-*.tmpl"}

var (
	// Center renders the full HTML shell with a simple centered body.
	Center = Layout{File: "layouts/center.tmpl"}

	// Equipment renders the full HTML shell with the inventory sidebar.
	Equipment = Layout{File: "layouts/equipment.tmpl", Partials: "partials/*.tmpl"}

	// Settings renders the full HTML shell with a GitHub-style settings sidebar.
	Settings = Layout{File: "layouts/settings.tmpl", Partials: "partials/*.tmpl"}

	// UserSettings renders the settings shell with the personal/account sidebar.
	UserSettings = Layout{File: "layouts/user-settings.tmpl", Partials: "partials/*.tmpl"}

	// Empty skips the HTML shell entirely; used for HTMX fragments.
	Empty = Layout{File: "layouts/empty.tmpl"}

	// Print renders a bare HTML shell for print-only pages (no nav, no chrome).
	Print = Layout{File: "layouts/print.tmpl"}

	// Login renders the full HTML shell with a login-focused body.
	Login = Layout{File: "layouts/login.tmpl"}
)
