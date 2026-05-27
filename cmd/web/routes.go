package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", http.RedirectHandler("/companies", http.StatusSeeOther))
	mux.HandleFunc("GET /media/{id}", app.handleHTML(app.getMedia))
	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))

	mux.HandleFunc("GET /companies", app.handleHTML(app.getCompanies))
	mux.HandleFunc("GET /companies/new", app.handleHTML(app.getCompaniesNew))
	mux.HandleFunc("POST /companies/new", app.handleHTML(app.postCompaniesNew))
	mux.HandleFunc("GET /companies/{company_id}", app.handleHTML(app.getSettingsCompany))
	mux.HandleFunc("POST /companies/{company_id}", app.handleHTML(app.postSettingsCompany))
	mux.HandleFunc("POST /companies/{company_id}/delete", app.handleHTML(app.postDeleteCompany))
	mux.HandleFunc("GET /companies/{company_id}/inventory", app.handleHTML(app.getInventory))
	mux.HandleFunc("GET /companies/{company_id}/inventory/new", app.handleHTML(app.getInventoryNew))
	mux.Handle("POST /companies/{company_id}/inventory/new", app.withCheckQuota(app.handleHTML(app.postInventoryNew)))
	mux.HandleFunc("GET /companies/{company_id}/inventory/{id}", app.handleHTML(app.getInventoryItem))
	mux.Handle("POST /companies/{company_id}/inventory/{id}", app.withCheckQuota(app.handleHTML(app.postInventoryItem)))
	mux.HandleFunc("POST /companies/{company_id}/inventory/{id}/delete", app.handleHTML(app.postDeleteInventoryItem))
	mux.HandleFunc("GET /companies/{company_id}/settings", app.handleHTML(app.getCompanySettings))
	mux.HandleFunc("POST /companies/{company_id}/settings", app.handleHTML(app.postCompanySettings))
	mux.HandleFunc("GET /companies/{company_id}/settings/equipment-categories", app.handleHTML(app.getEquipmentCategories))
	mux.HandleFunc("GET /companies/{company_id}/settings/equipment-categories/new", app.handleHTML(app.getEquipmentCategoryNew))
	mux.HandleFunc("POST /companies/{company_id}/settings/equipment-categories/new", app.handleHTML(app.postEquipmentCategoryNew))
	mux.HandleFunc("GET /companies/{company_id}/settings/equipment-categories/{id}", app.handleHTML(app.getEquipmentCategory))
	mux.HandleFunc("POST /companies/{company_id}/settings/equipment-categories/{id}", app.handleHTML(app.postEquipmentCategory))
	mux.HandleFunc("POST /companies/{company_id}/settings/equipment-categories/{id}/delete", app.handleHTML(app.postDeleteEquipmentCategory))

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
