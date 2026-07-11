// Package fragments lists the named template blocks available as HTMX fragments.
// Each constant matches the {{ define "name" }} block in the corresponding .tmpl file.
package fragments

// Fragment name constants for the HTMX partial templates.
const (
	EquipmentCategories    = "equipment-categories"
	EquipmentManufacturers = "equipment-manufacturers"
	EquipmentPartOf        = "equipment-part-of"
	WarehouseLocations     = "warehouse-locations"
	OrgCurrency            = "org-currency"
)
