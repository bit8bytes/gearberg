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
