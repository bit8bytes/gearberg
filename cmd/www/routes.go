package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/assets"
	"github.com/bit8bytes/gearberg/templates/pages"
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

	mux.HandleFunc("/", app.handleHTML(app.getLanding))

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

func (app *application) getLanding(w http.ResponseWriter, r *http.Request) *appError {
	data := app.newTemplateData(r)
	data.Data = struct {
		Year int
	}{
		Year: time.Now().Year(),
	}
	return app.render(w, r, http.StatusOK, pages.Landing, data)
}
