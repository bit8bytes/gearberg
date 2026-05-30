// Package layouts defines the available page layouts.
package layouts

// Layout groups a layout template file with the partials glob it requires.
// Partials is a glob pattern; leave empty when the layout needs no partials.
type Layout struct {
	File     string
	Partials string
}

var (
	// Landing renders the public landing page with a simple full-bleed layout.
	Landing = Layout{File: "layouts/landing.tmpl", Partials: "partials/*.tmpl"}

	// Center renders the full HTML shell with a simple centered body.
	Center = Layout{File: "layouts/center.tmpl"}

	// Inventory renders the full HTML shell with the inventory sidebar.
	Inventory = Layout{File: "layouts/inventory.tmpl", Partials: "partials/*.tmpl"}

	// Settings renders the full HTML shell with a GitHub-style settings sidebar.
	Settings = Layout{File: "layouts/settings.tmpl", Partials: "partials/*.tmpl"}

	// Empty skips the HTML shell entirely; used for HTMX fragments.
	Empty = Layout{File: "layouts/empty.tmpl"}
)
