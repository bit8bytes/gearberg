// Package assets package provides embedded static files for the application.
package assets

import (
	"embed"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

//go:embed dist
var files embed.FS

// ServeStaticFiles serves all embedded static files from the dist folder with
// gzip compression and a one-year Cache-Control header. ETags are set by the
// underlying FileServerFS, so browsers revalidate when the binary changes.
func ServeStaticFiles() http.Handler {
	fs := http.FileServerFS(files)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		fs.ServeHTTP(w, r)
	})

	gz, err := gzhttp.NewWrapper(gzhttp.MinSize(1024), gzhttp.CompressionLevel(6))
	if err != nil {
		// gzhttp.NewWrapper only errors on invalid options; panic is appropriate here.
		panic("assets: gzhttp.NewWrapper: " + err.Error())
	}
	return gz(inner)
}
