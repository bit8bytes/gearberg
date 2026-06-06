package main

import (
	"context"
	"time"

	gen "github.com/bit8bytes/gearberg/internal/api/gen"
	"github.com/bit8bytes/gearberg/internal/pagination"
)

func (app *application) GetHealthz(ctx context.Context) (*gen.HealthzResponse, error) {
	return &gen.HealthzResponse{
		Status:          "ok",
		AppRevision:     revision,
		DatabaseVersion: databaseVersion,
		Timestamp:       time.Now().UTC(),
	}, nil
}

func (app *application) ListInventory(ctx context.Context, params gen.ListInventoryParams) (*gen.InventoryListResponse, error) {
	page := params.Page.Or(1)
	f := pagination.Filters{Page: page, PageSize: 25}

	items, meta, err := app.services.inventory.GetFiltered(ctx, params.OrgID, params.Q.Or(""), "", f)
	if err != nil {
		return nil, err
	}

	out := make([]gen.InventoryItem, len(items))
	for i, item := range items {
		out[i] = gen.InventoryItem{
			ID:         item.ID,
			Name:       item.Name,
			Code:       item.Code,
			TotalStock: item.TotalStock,
			Type:       gen.InventoryItemType(item.Type.String()),
			UsageType:  gen.InventoryItemUsageType(item.UsageType.String()),
		}
		if item.CategoryName != "" {
			out[i].CategoryName = gen.NewOptString(item.CategoryName)
		}
	}

	return &gen.InventoryListResponse{
		Items:        out,
		CurrentPage:  meta.CurrentPage,
		LastPage:     meta.LastPage,
		TotalRecords: meta.TotalRecords,
	}, nil
}
