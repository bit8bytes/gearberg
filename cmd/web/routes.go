package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", http.RedirectHandler("/orgs", http.StatusSeeOther))
	mux.HandleFunc("GET /healthz", app.json.Handle(app.getHealthz))
	mux.HandleFunc("GET /media/{id}", app.html.Handle(app.getMedia))
	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))

	mux.HandleFunc("GET /orgs", app.html.Handle(app.getOrgs))
	mux.HandleFunc("GET /orgs/new", app.html.Handle(app.getOrgsNew))
	mux.HandleFunc("POST /orgs/new", app.html.Handle(app.postOrgsNew))
	mux.HandleFunc("GET /orgs/{org_id}", app.html.Handle(app.getSettingsOrg))
	mux.HandleFunc("POST /orgs/{org_id}", app.html.Handle(app.postSettingsOrg))
	mux.HandleFunc("POST /orgs/{org_id}/delete", app.html.Handle(app.postDeleteOrg))
	mux.HandleFunc("GET /orgs/{org_id}/inventory", app.html.Handle(app.getInventory))
	mux.HandleFunc("GET /orgs/{org_id}/inventory/new", app.html.Handle(app.getInventoryNew))
	mux.Handle("POST /orgs/{org_id}/inventory/new", app.withCheckQuota(app.html.Handle(app.postInventoryNew)))
	mux.HandleFunc("GET /orgs/{org_id}/inventory/{id}", app.html.Handle(app.getInventoryItem))
	mux.HandleFunc("GET /orgs/{org_id}/inventory/{id}/units", app.html.Handle(app.getInventoryUnits))
	mux.HandleFunc("POST /orgs/{org_id}/inventory/{id}/bulk", app.html.Handle(app.postInventoryItemBulk))
	mux.HandleFunc("POST /orgs/{org_id}/inventory/{id}/serialized", app.html.Handle(app.postInventoryItemSerialized))
	mux.HandleFunc("POST /orgs/{org_id}/inventory/{id}/units", app.html.Handle(app.postInventoryAddUnit))
	mux.HandleFunc("POST /orgs/{org_id}/inventory/{id}/units/{unit_id}", app.html.Handle(app.postInventoryUpdateUnit))
	mux.HandleFunc("POST /orgs/{org_id}/inventory/{id}/units/{unit_id}/delete", app.html.Handle(app.postDeleteInventoryUnit))
	mux.HandleFunc("POST /orgs/{org_id}/inventory/{id}/delete", app.html.Handle(app.postDeleteInventoryItem))
	mux.HandleFunc("GET /orgs/{org_id}/settings", app.html.Handle(app.getOrgSettings))
	mux.HandleFunc("POST /orgs/{org_id}/settings", app.html.Handle(app.postOrgSettings))
	mux.HandleFunc("GET /orgs/{org_id}/settings/equipment-categories", app.html.Handle(app.getEquipmentCategories))
	mux.HandleFunc("GET /orgs/{org_id}/settings/equipment-categories/new", app.html.Handle(app.getEquipmentCategoryNew))
	mux.HandleFunc("POST /orgs/{org_id}/settings/equipment-categories/new", app.html.Handle(app.postEquipmentCategoryNew))
	mux.HandleFunc("GET /orgs/{org_id}/settings/equipment-categories/{id}", app.html.Handle(app.getEquipmentCategory))
	mux.HandleFunc("POST /orgs/{org_id}/settings/equipment-categories/{id}", app.html.Handle(app.postEquipmentCategory))
	mux.HandleFunc("POST /orgs/{org_id}/settings/equipment-categories/{id}/delete", app.html.Handle(app.postDeleteEquipmentCategory))
	mux.HandleFunc("GET /orgs/{org_id}/settings/manufacturers", app.html.Handle(app.getManufacturers))
	mux.HandleFunc("GET /orgs/{org_id}/settings/manufacturers/new", app.html.Handle(app.getManufacturerNew))
	mux.HandleFunc("POST /orgs/{org_id}/settings/manufacturers/new", app.html.Handle(app.postManufacturerNew))
	mux.HandleFunc("GET /orgs/{org_id}/settings/manufacturers/{id}", app.html.Handle(app.getManufacturer))
	mux.HandleFunc("POST /orgs/{org_id}/settings/manufacturers/{id}", app.html.Handle(app.postManufacturer))
	mux.HandleFunc("POST /orgs/{org_id}/settings/manufacturers/{id}/delete", app.html.Handle(app.postDeleteManufacturer))

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
