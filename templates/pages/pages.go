// Package pages references all available application pages.
//
// Pages can be used by importing this package and referencing the page
// e.g. pages.CompaniesIndex.
//
// A page has two elements: the page file itself and the layout.
// Partials are owned by the layout, not the page.
package pages

import (
	"github.com/bit8bytes/gearberg/templates/layouts"
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
	// Companies is the company selector page.
	Companies = newPage("pages/companies/index.tmpl", layouts.Center)

	// CompaniesNew is the create-company form page.
	CompaniesNew = newPage("pages/companies/new.tmpl", layouts.Center)

	// CompanySettingsCompany is the General settings page — edit company name.
	CompanySettingsCompany = newPage("pages/companies/details.tmpl", layouts.Settings)

	// CompanySettingsConfig is the Billing & Tax settings page — currency, VAT rate, timezone.
	CompanySettingsConfig = newPage("pages/companies/settings.tmpl", layouts.Settings)

	// EquipmentCategoriesIndex is the Equipment Categories settings page.
	EquipmentCategoriesIndex = newPage("pages/companies/categories/index.tmpl", layouts.Settings)

	// EquipmentCategoriesNew is the create-category form page.
	EquipmentCategoriesNew = newPage("pages/companies/categories/new.tmpl", layouts.Center)

	// EquipmentCategoriesDetail is the category detail/edit page.
	EquipmentCategoriesDetail = newPage("pages/companies/categories/detail.tmpl", layouts.Center)

	// Inventory is the inventory page.
	Inventory = newPage("pages/inventory/index.tmpl", layouts.Inventory)

	// NotFound is the 404 page.
	NotFound = newPage("pages/not-found.tmpl", layouts.Center)

	// Error is the generic error page.
	Error = newPage("pages/error.tmpl", layouts.Center)
)
