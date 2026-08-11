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

// Package fragments lists the named template blocks available as HTMX fragments.
// Each constant matches the {{ define "name" }} block in the corresponding .tmpl file.
package fragments

// Fragment name constants for the HTMX partial templates.
const (
	EquipmentCategories    = "equipment-categories"
	EquipmentManufacturers = "equipment-manufacturers"
	EquipmentPartOf        = "equipment-part-of"
	EquipmentSearch        = "equipment-search"
	WarehouseLocations     = "warehouse-locations"
	OrgCurrency            = "org-currency"
	PasswordValidation     = "password-validation"
)
