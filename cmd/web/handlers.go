package main

import (
	"net/http"
	"time"

	"github.com/bit8bytes/gearberg/internal/httperr"
)

type response struct {
	Status          string    `json:"status"`
	AppRevision     string    `json:"app_revision"`
	DatabaseVersion string    `json:"database_version"`
	Timestamp       time.Time `json:"timestamp"`
}

func (app *application) getHealthz(w http.ResponseWriter, r *http.Request) *httperr.Error {
	return app.json.Write(r.Context(), w, http.StatusOK, response{
		Status:          "ok",
		AppRevision:     revision,
		DatabaseVersion: databaseVersion,
		Timestamp:       time.Now().UTC(),
	}, nil)
}
