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
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gen "github.com/bit8bytes/gearberg/internal/api/gen"
	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/pagination"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

func (app *application) getNotFound(w http.ResponseWriter, r *http.Request) *httperr.Error {
	return app.html.Render(w, r, http.StatusOK, pages.NotFound, app.html.TemplateData(r))
}

func (app *application) getForbidden(w http.ResponseWriter, r *http.Request) *httperr.Error {
	return app.html.Render(w, r, http.StatusOK, pages.Forbidden, app.html.TemplateData(r))
}

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

	items, meta, err := app.services.equipment.GetFiltered(ctx, params.OrgID, params.Q.Or(""), "", f)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	out := make([]gen.EquipmentItem, len(items))
	for i, item := range items {
		out[i] = gen.EquipmentItem{
			ID:         item.ID,
			Name:       item.Name,
			TotalStock: item.TotalStock,
			Type:       gen.EquipmentItemType(item.TrackingType.String()),
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
