package main

import (
	"fmt"
	"net/http"

	gen "github.com/bit8bytes/gearberg/internal/api/gen"
	"github.com/bit8bytes/gearberg/internal/assets"
)

func (app *application) routes() (http.Handler, error) {
	mux := http.NewServeMux()

	mux.Handle("GET /{$}", http.RedirectHandler("/orgs", http.StatusSeeOther))
	apiServer, err := gen.NewServer(app, gen.WithPathPrefix("/api/v1"))
	if err != nil {
		return nil, fmt.Errorf("routes: new api server: %w", err)
	}
	mux.Handle("/api/v1/", apiServer)
	mux.HandleFunc("GET /media/{id}", app.html.Handle(app.getMedia))
	mux.HandleFunc("GET /image-proxy", app.getImageProxy)
	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))

	mux.HandleFunc("GET /orgs", app.html.Handle(app.getOrgs))
	mux.HandleFunc("GET /orgs/new", app.html.Handle(app.getOrgsNew))
	mux.HandleFunc("POST /orgs/new", app.html.Handle(app.postOrgsNew))
	mux.HandleFunc("GET /orgs/{org_id}", app.html.Handle(app.getSettingsOrg))
	mux.HandleFunc("POST /orgs/{org_id}", app.html.Handle(app.postSettingsOrg))
	mux.HandleFunc("POST /orgs/{org_id}/delete", app.html.Handle(app.postDeleteOrg))
	mux.HandleFunc("GET /orgs/{org_id}/equipment", app.html.Handle(app.getEquipment))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/export", app.html.Handle(app.getEquipmentExport))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/print", app.html.Handle(app.getEquipmentPrint))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/import", app.html.Handle(app.getEquipmentImport))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/import", app.html.Handle(app.postEquipmentImport))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/import/confirm", app.html.Handle(app.postEquipmentImportConfirm))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/new", app.html.Handle(app.getEquipmentNew))
	mux.Handle("POST /orgs/{org_id}/equipment/new", app.withCheckQuota(app.html.Handle(app.postEquipmentNew)))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/{id}", app.html.Handle(app.getEquipmentItem))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/{id}/units", app.html.Handle(app.getEquipmentUnits))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/{id}/pricing", app.html.Handle(app.getEquipmentItemPricing))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/pricing", app.html.Handle(app.postEquipmentItemPricing))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/{id}/properties", app.html.Handle(app.getEquipmentItemProperties))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/details", app.html.Handle(app.postEquipmentItemDetails))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/properties", app.html.Handle(app.postEquipmentItemProperties))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/units", app.html.Handle(app.postEquipmentAddUnit))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/units/{unit_id}", app.html.Handle(app.postEquipmentUpdateUnit))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/{id}/units/{unit_id}/qr", app.html.Handle(app.getEquipmentUnitQR))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/units/{unit_id}/delete", app.html.Handle(app.postDeleteEquipmentUnit))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/archive", app.html.Handle(app.postArchiveEquipmentItem))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/delete", app.html.Handle(app.postDeleteEquipmentItem))
	mux.HandleFunc("GET /orgs/{org_id}/equipment/{id}/content", app.html.Handle(app.getEquipmentContent))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/content", app.html.Handle(app.postEquipmentAssignContent))
	mux.HandleFunc("POST /orgs/{org_id}/equipment/{id}/content/{content_id}/delete", app.html.Handle(app.postEquipmentRemoveContent))
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
	mux.HandleFunc("GET /orgs/{org_id}/settings/locations", app.html.Handle(app.getLocations))
	mux.HandleFunc("GET /orgs/{org_id}/settings/locations/new", app.html.Handle(app.getLocationNew))
	mux.HandleFunc("POST /orgs/{org_id}/settings/locations/new", app.html.Handle(app.postLocationNew))
	mux.HandleFunc("GET /orgs/{org_id}/settings/locations/{id}", app.html.Handle(app.getLocation))
	mux.HandleFunc("POST /orgs/{org_id}/settings/locations/{id}", app.html.Handle(app.postLocation))
	mux.HandleFunc("POST /orgs/{org_id}/settings/locations/{id}/delete", app.html.Handle(app.postDeleteLocation))

	antiCSRF := http.NewCrossOriginProtection()
	logRequest := newRequestLogger(app.logger)
	recoverPanic := newPanicRecoverer(app.logger)

	return withTrace(
		withNonce(
			recoverPanic.handler(
				logRequest.handler(
					withSecurityHeaders(
						withMaxBodySize(
							antiCSRF.Handler(mux))))))), nil
}
