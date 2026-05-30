package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", http.RedirectHandler("/companies", http.StatusSeeOther))
	mux.HandleFunc("GET /media/{id}", app.html.Handle(app.getMedia))
	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))

	mux.HandleFunc("GET /companies", app.html.Handle(app.getCompanies))
	mux.HandleFunc("GET /companies/new", app.html.Handle(app.getCompaniesNew))
	mux.HandleFunc("POST /companies/new", app.html.Handle(app.postCompaniesNew))
	mux.HandleFunc("GET /companies/{company_id}", app.html.Handle(app.getSettingsCompany))
	mux.HandleFunc("POST /companies/{company_id}", app.html.Handle(app.postSettingsCompany))
	mux.HandleFunc("POST /companies/{company_id}/delete", app.html.Handle(app.postDeleteCompany))
	mux.HandleFunc("GET /companies/{company_id}/inventory", app.html.Handle(app.getInventory))
	mux.HandleFunc("GET /companies/{company_id}/inventory/new", app.html.Handle(app.getInventoryNew))
	mux.Handle("POST /companies/{company_id}/inventory/new", app.withCheckQuota(app.html.Handle(app.postInventoryNew)))
	mux.HandleFunc("GET /companies/{company_id}/inventory/{id}", app.html.Handle(app.getInventoryItem))
	mux.Handle("POST /companies/{company_id}/inventory/{id}", app.withCheckQuota(app.html.Handle(app.postInventoryItem)))
	mux.HandleFunc("POST /companies/{company_id}/inventory/{id}/delete", app.html.Handle(app.postDeleteInventoryItem))
	mux.HandleFunc("GET /companies/{company_id}/settings", app.html.Handle(app.getCompanySettings))
	mux.HandleFunc("POST /companies/{company_id}/settings", app.html.Handle(app.postCompanySettings))
	mux.HandleFunc("GET /companies/{company_id}/settings/equipment-categories", app.html.Handle(app.getEquipmentCategories))
	mux.HandleFunc("GET /companies/{company_id}/settings/equipment-categories/new", app.html.Handle(app.getEquipmentCategoryNew))
	mux.HandleFunc("POST /companies/{company_id}/settings/equipment-categories/new", app.html.Handle(app.postEquipmentCategoryNew))
	mux.HandleFunc("GET /companies/{company_id}/settings/equipment-categories/{id}", app.html.Handle(app.getEquipmentCategory))
	mux.HandleFunc("POST /companies/{company_id}/settings/equipment-categories/{id}", app.html.Handle(app.postEquipmentCategory))
	mux.HandleFunc("POST /companies/{company_id}/settings/equipment-categories/{id}/delete", app.html.Handle(app.postDeleteEquipmentCategory))

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
