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
package equipment

import (
	"github.com/bit8bytes/gearberg/internal/database"
	genequip "github.com/bit8bytes/gearberg/internal/database/queries/gen/equipment"
)

func pricingFromListRow(row genequip.ListRow) Pricing {
	return Pricing{
		PurchasePrice: database.NullAs[Cents](row.ResalePrice),
		RentalPrice:   database.NullAs[Cents](row.RentalPrice),
	}
}

func propertiesFromCreateRow(row genequip.CreateRow) Properties {
	return Properties{
		Weight:    database.NullAs[Grams](row.WeightG),
		Width:     database.NullAs[Millimeters](row.WidthMm),
		Height:    database.NullAs[Millimeters](row.HeightMm),
		Depth:     database.NullAs[Millimeters](row.DepthMm),
		Power:     database.NullAs[Milliwatts](row.PowerMw),
		Current:   database.NullAs[Milliamps](row.CurrentMa),
		Voltage:   database.NullAs[Millivolts](row.VoltageMv),
		WireGauge: database.NullAs[WireGauge](row.WireGaugeMm2X100),
	}
}

func pricingFromCreateRow(row genequip.CreateRow) Pricing {
	return Pricing{
		PurchasePrice: database.NullAs[Cents](row.ResalePrice),
		RentalPrice:   database.NullAs[Cents](row.RentalPrice),
	}
}

func propertiesFromGetByIDRow(row genequip.GetByIDRow) Properties {
	return Properties{
		Weight:    database.NullAs[Grams](row.WeightG),
		Width:     database.NullAs[Millimeters](row.WidthMm),
		Height:    database.NullAs[Millimeters](row.HeightMm),
		Depth:     database.NullAs[Millimeters](row.DepthMm),
		Power:     database.NullAs[Milliwatts](row.PowerMw),
		Current:   database.NullAs[Milliamps](row.CurrentMa),
		Voltage:   database.NullAs[Millivolts](row.VoltageMv),
		WireGauge: database.NullAs[WireGauge](row.WireGaugeMm2X100),
	}
}

func pricingFromGetByIDRow(row genequip.GetByIDRow) Pricing {
	return Pricing{
		PurchasePrice: database.NullAs[Cents](row.ResalePrice),
		RentalPrice:   database.NullAs[Cents](row.RentalPrice),
	}
}

func propertiesFromListRow(row genequip.ListRow) Properties {
	return Properties{
		Weight:    database.NullAs[Grams](row.WeightG),
		Width:     database.NullAs[Millimeters](row.WidthMm),
		Height:    database.NullAs[Millimeters](row.HeightMm),
		Depth:     database.NullAs[Millimeters](row.DepthMm),
		Power:     database.NullAs[Milliwatts](row.PowerMw),
		Current:   database.NullAs[Milliamps](row.CurrentMa),
		Voltage:   database.NullAs[Millivolts](row.VoltageMv),
		WireGauge: database.NullAs[WireGauge](row.WireGaugeMm2X100),
	}
}
