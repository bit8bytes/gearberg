package main

import (
	"net/http"

	"github.com/bit8bytes/gearberg/internal/httperr"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

type dashboardData struct {
	OrgID string
}

func (app *application) getDashboard(w http.ResponseWriter, r *http.Request) *httperr.Error {
	tmplData := app.html.TemplateData(r)
	tmplData.Data = dashboardData{
		OrgID: r.PathValue("org_id"),
	}
	return app.html.Render(w, r, http.StatusOK, pages.Dashboard, tmplData)
}
