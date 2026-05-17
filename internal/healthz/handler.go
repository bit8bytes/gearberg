// Package healthz provides the health check HTTP handler.
package healthz

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler handles health check requests.
type Handler struct {
	appRevision     string
	databaseVersion string
}

type response struct {
	Status          string    `json:"status"`
	AppRevision     string    `json:"app_revision"`
	DatabaseVersion string    `json:"database_version"`
	Timestamp       time.Time `json:"timestamp"`
}

// NewHandler returns a new Handler with the given build revision string.
func NewHandler(appRevision string, databaseVersion string) *Handler {
	return &Handler{
		appRevision:     appRevision,
		databaseVersion: databaseVersion,
	}
}

// Routes registers the health check endpoint on the given mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.get)
}

func (h *Handler) get(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response{
		Status:          "ok",
		AppRevision:     h.appRevision,
		DatabaseVersion: h.databaseVersion,
		Timestamp:       time.Now().UTC(),
	}); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
