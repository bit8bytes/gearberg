// Package templates holds the HTML templates embedded into the binary so the
// application can be deployed as a single self-contained executable without
// requiring template files to exist on the host filesystem.
package templates

import "embed"

// EmbedFS contains all template directories embedded at compile time.
//
//go:embed components layouts pages partials
var EmbedFS embed.FS
