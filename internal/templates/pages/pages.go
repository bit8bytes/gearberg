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

// Package pages references all available application pages.
//
// Pages can be used by importing this package and referencing the page
// e.g. pages.Name.
//
// A page has two elements: the page file itself and the layout.
// Partials are owned by the layout, not the page.
package pages

import (
	"github.com/bit8bytes/gearberg/internal/templates/layouts"
)

// Page holds the template file path and the layout used to render it.
type Page struct {
	Layout layouts.Layout
	File   string
}

// All is populated automatically as each page var is declared via newPage.
var All []Page

func newPage(file string, layout layouts.Layout) Page {
	p := Page{File: file, Layout: layout}
	All = append(All, p)
	return p
}

var (
	// Landing is the public landing page (English).
	Landing = newPage("pages/landing.tmpl", layouts.Landing)

	// LandingDE is the German landing page.
	LandingDE = newPage("pages/landing.de.tmpl", layouts.Landing)
)

// Docs is the public docs page.
var Docs = newPage("pages/docs.tmpl", layouts.Docs)

// Web app pages.
var (
	// Orgs is the org selector page.
	Orgs = newPage("pages/orgs/index.tmpl", layouts.Center)
	// OrgPicker is the full-screen org selection page shown after login with multiple orgs.
	OrgPicker = newPage("pages/orgs/picker.tmpl", layouts.Login)
	// OrgsNew is the create-org form page.
	OrgsNew = newPage("pages/orgs/new.tmpl", layouts.Center)
	// OrgSettingsDetails is the General settings page — edit org name.
	OrgSettingsDetails = newPage("pages/orgs/details.tmpl", layouts.Settings)
	// OrgSettings is the Billing & Tax settings page — currency, VAT rate, timezone.
	OrgSettings = newPage("pages/orgs/localization.tmpl", layouts.Settings)

	Dashboard = newPage("pages/dashboard.tmpl", layouts.Equipment)

	// EquipmentCategoriesIndex is the Equipment Categories settings page.
	EquipmentCategoriesIndex = newPage("pages/orgs/categories/index.tmpl", layouts.Settings)
	// EquipmentCategoriesNew is the create-category form page.
	EquipmentCategoriesNew = newPage("pages/orgs/categories/new.tmpl", layouts.Center)
	// EquipmentCategoriesDetail is the category detail/edit page.
	EquipmentCategoriesDetail = newPage("pages/orgs/categories/detail.tmpl", layouts.Center)
	// ManufacturersIndex is the Manufacturers settings page.
	ManufacturersIndex = newPage("pages/orgs/manufacturers/index.tmpl", layouts.Settings)
	// ManufacturersNew is the create-manufacturer form page.
	ManufacturersNew = newPage("pages/orgs/manufacturers/new.tmpl", layouts.Center)
	// ManufacturersDetail is the manufacturer detail/edit page.
	ManufacturersDetail = newPage("pages/orgs/manufacturers/detail.tmpl", layouts.Center)
	// LocationsIndex is the Locations settings page.
	LocationsIndex = newPage("pages/orgs/locations/index.tmpl", layouts.Settings)
	// LocationsNew is the create-location form page.
	LocationsNew = newPage("pages/orgs/locations/new.tmpl", layouts.Center)
	// LocationsDetail is the location detail/edit page.
	LocationsDetail = newPage("pages/orgs/locations/detail.tmpl", layouts.Center)
	// Equipment is the equipment list page.
	Equipment = newPage("pages/equipment/index.tmpl", layouts.Equipment)
	// EquipmentNew is the unified create-item form page (bulk or serialized).
	EquipmentNew = newPage("pages/equipment/new.tmpl", layouts.Center)
	// EquipmentDetail is the item details tab page.
	EquipmentDetail = newPage("pages/equipment/detail.tmpl", layouts.Equipment)
	// EquipmentPricing is the item pricing tab page.
	EquipmentPricing = newPage("pages/equipment/pricing.tmpl", layouts.Equipment)
	// EquipmentProperties is the item properties tab page.
	EquipmentProperties = newPage("pages/equipment/properties.tmpl", layouts.Equipment)
	// EquipmentUnits is the unit management tab page for a serialized item.
	EquipmentUnits = newPage("pages/equipment/units.tmpl", layouts.Equipment)
	// EquipmentContent is the content management tab page for an item with has_content=true.
	EquipmentContent = newPage("pages/equipment/content.tmpl", layouts.Equipment)
	// EquipmentImport is the CSV upload form page.
	EquipmentImport = newPage("pages/equipment/import.tmpl", layouts.Center)
	// EquipmentImportPreview is the staging review and confirm page.
	EquipmentImportPreview = newPage("pages/equipment/import-preview.tmpl", layouts.Equipment)
	// EquipmentPrint is the bare print view of the equipment list.
	EquipmentPrint = newPage("pages/equipment/print.tmpl", layouts.Print)

	// ImportReview is the staged-row review and confirm page.
	ImportReview = newPage("pages/import/review.tmpl", layouts.Center)

	// Imprint is the legal imprint page.
	Imprint = newPage("pages/imprint.tmpl", layouts.Landing)
	// Privacy is the privacy policy page.
	Privacy = newPage("pages/privacy.tmpl", layouts.Landing)

	// NotFound is the 404 page.
	NotFound = newPage("pages/not-found.tmpl", layouts.Center)
	// Forbidden is 403 page.
	Forbidden = newPage("pages/forbidden.tmpl", layouts.Center)
	// Error is the generic error page.
	Error = newPage("pages/error.tmpl", layouts.Center)

	// SignIn is the sign-in page.
	SignIn = newPage("pages/login/signin.tmpl", layouts.Login)
	// SignUp is the account registration page.
	SignUp = newPage("pages/login/signup.tmpl", layouts.Login)
	// ForgotPassword is the forgot-password request page.
	ForgotPassword = newPage("pages/login/forgot-password.tmpl", layouts.Login)
	// ForgotPasswordSuccess is shown after a reset email has been sent.
	ForgotPasswordSuccess = newPage("pages/login/forgot-password-success.tmpl", layouts.Login)
	// ResetPassword is the page where users submit their new password.
	ResetPassword = newPage("pages/login/reset-password.tmpl", layouts.Login)

	// SettingsAccount is the account settings page.
	SettingsAccount = newPage("pages/settings/account.tmpl", layouts.UserSettings)

	// SettingsOrganizations is the org list/selector page within settings.
	SettingsOrganizations = newPage("pages/orgs/index.tmpl", layouts.UserSettings)
)
