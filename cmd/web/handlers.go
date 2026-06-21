package main

import (
	"context"
	"fmt"
	"time"

	gen "github.com/bit8bytes/gearberg/internal/api/gen"
	"github.com/bit8bytes/gearberg/internal/pagination"
)

func (app *application) GetHealthz(_ context.Context) (*gen.HealthzResponse, error) {
	return &gen.HealthzResponse{
		Status:          "ok",
		AppRevision:     revision,
		DatabaseVersion: databaseVersion,
		Timestamp:       time.Now().UTC(),
	}, nil
}

func (app *application) ListEquipment(ctx context.Context, params gen.ListEquipmentParams) (*gen.EquipmentListResponse, error) {
	page := params.Page.Or(1)
	f := pagination.Filters{Page: page, PageSize: 25}

	items, meta, err := app.services.equipment.GetFiltered(ctx, params.OrgID, params.Q.Or(""), "", false, f)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	out := make([]gen.EquipmentItem, len(items))
	for i, item := range items {
		out[i] = gen.EquipmentItem{
			ID:         item.ID,
			Name:       item.Name,
			TotalStock: item.TotalStock,
			Type:       gen.EquipmentItemType(item.Type.String()),
			UsageType:  gen.EquipmentItemUsageType(item.UsageType.String()),
		}
		if item.CategoryName != "" {
			out[i].CategoryName = gen.NewOptString(item.CategoryName)
		}
	}

	return &gen.EquipmentListResponse{
		Items:        out,
		CurrentPage:  meta.CurrentPage,
		LastPage:     meta.LastPage,
		TotalRecords: meta.TotalRecords,
	}, nil
}
