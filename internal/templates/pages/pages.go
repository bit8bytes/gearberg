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

	// Inventory is the inventory list page.
	Inventory = newPage("pages/inventory/index.tmpl", layouts.Inventory)

	// InventoryNew is the unified create-item form page (bulk or serialized).
	InventoryNew = newPage("pages/inventory/new.tmpl", layouts.Center)

	// InventoryDetailBulk is the bulk item detail/edit page.
	InventoryDetailBulk = newPage("pages/inventory/detail-bulk.tmpl", layouts.Center)

	// InventoryDetailSerialized is the serialized item detail/edit page.
	InventoryDetailSerialized = newPage("pages/inventory/detail-serialized.tmpl", layouts.Center)

	// InventoryUnits is the unit management page for a serialized item.
	InventoryUnits = newPage("pages/inventory/units.tmpl", layouts.Center)

	// InventoryImport is the CSV upload form page.
	InventoryImport = newPage("pages/inventory/import.tmpl", layouts.Center)

	// InventoryImportPreview is the staging review and confirm page.
	InventoryImportPreview = newPage("pages/inventory/import-preview.tmpl", layouts.Inventory)

	// InventoryPrint is the bare print view of the inventory list.
	InventoryPrint = newPage("pages/inventory/print.tmpl", layouts.Print)

	// Imprint is the legal imprint page.
	Imprint = newPage("pages/imprint.tmpl", layouts.Landing)

	// Privacy is the privacy policy page.
	Privacy = newPage("pages/privacy.tmpl", layouts.Landing)

	// NotFound is the 404 page.
	NotFound = newPage("pages/not-found.tmpl", layouts.Center)

	// Error is the generic error page.
	Error = newPage("pages/error.tmpl", layouts.Center)
)
