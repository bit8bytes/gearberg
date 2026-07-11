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
