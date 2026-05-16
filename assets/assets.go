// Package assets package provides embedded static files for the application.
package assets

import (
	"embed"
	"net/http"
)

//go:embed dist
var files embed.FS

// ServeStaticFiles serves all embedded static [files] from dist folder.
func ServeStaticFiles() http.Handler {
	fs := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}
