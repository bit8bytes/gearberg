package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/internal/assets"
	"github.com/bit8bytes/gearberg/internal/httperr"
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
