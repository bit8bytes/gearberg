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

// Pages.
var (
	// Landing is the public landing page.
	Landing = newPage("pages/landing.tmpl", layouts.Landing)

	// Orgs is the org selector page.
	Orgs = newPage("pages/orgs/index.tmpl", layouts.Center)
	// OrgsNew is the create-org form page.
	OrgsNew = newPage("pages/orgs/new.tmpl", layouts.Center)
	// OrgSettingsDetails is the General settings page — edit org name.
	OrgSettingsDetails = newPage("pages/orgs/details.tmpl", layouts.Settings)
	// OrgSettings is the Billing & Tax settings page — currency, VAT rate, timezone.
	OrgSettings = newPage("pages/orgs/settings.tmpl", layouts.Settings)
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
	SettingsAccount = newPage("pages/settings/account.tmpl", layouts.Center)
)
