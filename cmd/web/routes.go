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
package main

import (
	"fmt"
	"net/http"

	gen "github.com/bit8bytes/gearberg/internal/api/gen"
	"github.com/bit8bytes/gearberg/internal/assets"
)

func (app *application) routes() (http.Handler, error) {
	mux := http.NewServeMux()
	apiServer, err := gen.NewServer(app, gen.WithPathPrefix("/api/v1"))
	if err != nil {
		return nil, fmt.Errorf("routes: new api server: %w", err)
	}

	mux.HandleFunc("/", app.html.Handle(app.getNotFound))
	mux.Handle("GET /forbidden", app.html.Handle(app.getForbidden))

	mux.Handle("GET /{$}", http.RedirectHandler("/settings/organizations", http.StatusSeeOther))
	mux.Handle("GET /dist/", assets.ServeStaticFiles())
	mux.Handle("GET /favicon.ico", http.RedirectHandler("/dist/images/favicon.ico", http.StatusMovedPermanently))

	// OIDC routes for multiple providers.
	for name := range app.oidcProviders {
		mux.Handle("GET /auth/oidc/"+name, app.withGuest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.getAuthOIDC(w, r, name)
		})))
		mux.Handle("GET /auth/oidc/"+name+"/callback", app.withGuest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.getAuthOIDCCallback(w, r, name)
		})))
	}

	// Login related routes that are only accessible if the user is not logged in.
	mux.Handle("GET /signin", app.withGuest(app.html.Handle(app.getSignIn)))
	mux.Handle("POST /signin", app.withGuest(app.html.Handle(app.postSignIn)))
	mux.Handle("GET /signup", app.withGuest(app.html.Handle(app.getSignUp)))
	mux.Handle("POST /signup", app.withGuest(app.html.Handle(app.postSignUp)))
	mux.Handle("GET /forgot-password", app.withGuest(app.html.Handle(app.getForgotPassword)))
	mux.Handle("POST /forgot-password", app.withGuest(app.html.Handle(app.postForgotPassword)))
	mux.Handle("GET /forgot-password/success", app.withGuest(app.html.Handle(app.getForgotPasswordSuccess)))
	mux.Handle("GET /reset-password", app.withGuest(app.html.Handle(app.getResetPassword)))
	mux.Handle("POST /reset-password", app.withGuest(app.html.Handle(app.postResetPassword)))
	mux.Handle("POST /validate/password", app.html.Handle(app.postValidatePassword))

	// TODO: withLogin needs to be replaced with api specific bearer tokens.
	mux.Handle("/api/v1/", app.withLogin(apiServer))

	// Signout must only be possible for logged in users.
	mux.Handle("POST /signout", app.withLogin(app.html.Handle(app.postSignOut)))

	mux.Handle("GET /media/{id}", app.withLogin(app.html.Handle(app.getMedia)))
	mux.Handle("GET /image-proxy", app.withLogin(http.HandlerFunc(app.getImageProxy)))

	// Org related actions. The account only needs to be logged in via [app.withLogin].
	mux.Handle("GET /orgs", app.withLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/settings/organizations", http.StatusSeeOther)
	})))
	mux.Handle("GET /orgs/new", app.withLogin(app.html.Handle(app.getOrgsNew)))
	mux.Handle("POST /orgs/new", app.withLogin(app.html.Handle(app.postOrgsNew)))

	// For any route that includes org_id, the permissions of the account must be checked via [app.withPermission]
	mux.Handle("DELETE /orgs/{org_id}", app.withLogin(app.withPermission(app.html.Handle(app.deleteOrg))))
	mux.Handle("GET /orgs/{org_id}", app.withLogin(app.withPermission(app.html.Handle(app.getSettingsOrg))))
	mux.Handle("POST /orgs/{org_id}", app.withLogin(app.withPermission(app.html.Handle(app.postSettingsOrg))))

	// Equipment
	mux.Handle("GET /orgs/{org_id}/equipment", app.withLogin(app.withPermission(app.html.Handle(app.getEquipment))))
	mux.Handle("GET /orgs/{org_id}/equipment/print", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentPrint))))
	mux.Handle("GET /orgs/{org_id}/equipment/export", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentExport))))
	mux.Handle("GET /orgs/{org_id}/equipment/import", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentImport))))
	mux.Handle("GET /orgs/{org_id}/equipment/import/template", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentImportTemplate))))
	mux.Handle("POST /orgs/{org_id}/equipment/import", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentImport))))
	mux.Handle("POST /orgs/{org_id}/equipment/import/confirm", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentImportConfirm))))
	mux.Handle("GET /orgs/{org_id}/equipment/new", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentNew))))
	mux.Handle("POST /orgs/{org_id}/equipment/new", app.withLogin(app.withPermission(app.withCheckQuota(app.html.Handle(app.postEquipmentNew)))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentItem))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/units", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentUnits))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/pricing", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentItemPricing))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/pricing", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentItemPricing))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/properties", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentItemProperties))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/details", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentItemDetails))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/properties", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentItemProperties))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/units", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentAddUnit))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/units/bulk-inspect", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentBulkUpdateInspection))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/units/{unit_id}", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentUpdateUnit))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/units/{unit_id}/qr", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentUnitQR))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/units/{unit_id}/barcode", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentUnitBarcode))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/units/{unit_id}/delete", app.withLogin(app.withPermission(app.html.Handle(app.postDeleteEquipmentUnit))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/delete", app.withLogin(app.withPermission(app.html.Handle(app.postDeleteEquipmentItem))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/content", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentContent))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/content", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentAssignContent))))
	mux.Handle("POST /orgs/{org_id}/equipment/{id}/content/{content_id}/delete", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentRemoveContent))))
	mux.Handle("GET /orgs/{org_id}/equipment/{id}/part-of", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentPartOfFragment))))
	mux.Handle("GET /orgs/{org_id}/currency", app.withLogin(app.withPermission(app.html.Handle(app.getOrgCurrencyFragment))))
	mux.Handle("GET /orgs/{org_id}/equipment-categories", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentCategoriesFragment))))
	mux.Handle("GET /orgs/{org_id}/equipment-manufacturers", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentManufacturersFragment))))
	mux.Handle("GET /orgs/{org_id}/warehouse-locations", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentLocationsFragment))))

	// Account fragments
	mux.Handle("GET /account/header", app.withLogin(app.html.Handle(app.getAccountHeaderFragment)))

	// Account Settings
	mux.Handle("GET /settings/account", app.withLogin(app.html.Handle(app.getAccount)))
	mux.Handle("DELETE /settings/account", app.withLogin(app.html.Handle(app.deleteAccount)))
	mux.Handle("GET /settings/organizations", app.withLogin(app.html.Handle(app.getOrgs)))

	// Org related settings
	mux.Handle("GET /orgs/{org_id}/settings", app.withLogin(app.withPermission(app.html.Handle(app.getOrgSettings))))
	mux.Handle("POST /orgs/{org_id}/settings", app.withLogin(app.withPermission(app.html.Handle(app.postOrgSettings))))
	mux.Handle("GET /orgs/{org_id}/settings/equipment-categories", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentCategories))))
	mux.Handle("GET /orgs/{org_id}/settings/equipment-categories/new", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentCategoryNew))))
	mux.Handle("POST /orgs/{org_id}/settings/equipment-categories/new", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentCategoryNew))))
	mux.Handle("GET /orgs/{org_id}/settings/equipment-categories/{id}", app.withLogin(app.withPermission(app.html.Handle(app.getEquipmentCategory))))
	mux.Handle("POST /orgs/{org_id}/settings/equipment-categories/{id}", app.withLogin(app.withPermission(app.html.Handle(app.postEquipmentCategory))))
	mux.Handle("POST /orgs/{org_id}/settings/equipment-categories/{id}/delete", app.withLogin(app.withPermission(app.html.Handle(app.postDeleteEquipmentCategory))))
	mux.Handle("GET /orgs/{org_id}/settings/manufacturers", app.withLogin(app.withPermission(app.html.Handle(app.getManufacturers))))
	mux.Handle("GET /orgs/{org_id}/settings/manufacturers/new", app.withLogin(app.withPermission(app.html.Handle(app.getManufacturerNew))))
	mux.Handle("POST /orgs/{org_id}/settings/manufacturers/new", app.withLogin(app.withPermission(app.html.Handle(app.postManufacturerNew))))
	mux.Handle("GET /orgs/{org_id}/settings/manufacturers/{id}", app.withLogin(app.withPermission(app.html.Handle(app.getManufacturer))))
	mux.Handle("POST /orgs/{org_id}/settings/manufacturers/{id}", app.withLogin(app.withPermission(app.html.Handle(app.postManufacturer))))
	mux.Handle("POST /orgs/{org_id}/settings/manufacturers/{id}/delete", app.withLogin(app.withPermission(app.html.Handle(app.postDeleteManufacturer))))
	mux.Handle("GET /orgs/{org_id}/settings/locations", app.withLogin(app.withPermission(app.html.Handle(app.getLocations))))
	mux.Handle("GET /orgs/{org_id}/settings/locations/new", app.withLogin(app.withPermission(app.html.Handle(app.getLocationNew))))
	mux.Handle("POST /orgs/{org_id}/settings/locations/new", app.withLogin(app.withPermission(app.html.Handle(app.postLocationNew))))
	mux.Handle("GET /orgs/{org_id}/settings/locations/{id}", app.withLogin(app.withPermission(app.html.Handle(app.getLocation))))
	mux.Handle("POST /orgs/{org_id}/settings/locations/{id}", app.withLogin(app.withPermission(app.html.Handle(app.postLocation))))
	mux.Handle("POST /orgs/{org_id}/settings/locations/{id}/delete", app.withLogin(app.withPermission(app.html.Handle(app.postDeleteLocation))))

	antiCSRF := http.NewCrossOriginProtection()
	logRequest := newRequestLogger(app.logger)
	recoverPanic := newPanicRecoverer(app.logger)

	return withTrace(
		withNonce(
			recoverPanic.handler(
				logRequest.handler(
					withSecurityHeaders(
						withMaxBodySize(
							antiCSRF.Handler(
								app.session.LoadAndSave(mux)))))))), nil
}
