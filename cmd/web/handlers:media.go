package main

import (
	"io"
	"net/http"
)

func (app *application) getMedia(w http.ResponseWriter, r *http.Request) *appError {
	id := r.PathValue("id")

	obj, err := app.services.storageManager.Get(r.Context(), id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Media not found.",
			Code:    http.StatusNotFound,
		}
	}

	rc, err := app.services.storageManager.Open(r.Context(), id)
	if err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to open media.",
			Code:    http.StatusInternalServerError,
		}
	}
	defer func(rc io.ReadCloser) {
		if err := rc.Close(); err != nil {
			app.logger.Warn("Failed to close media reader", "error", err)
		}
	}(rc)

	w.Header().Set("Content-Type", obj.ContentType)
	if _, err := io.Copy(w, rc); err != nil {
		return &appError{
			Error:   err,
			Message: "Failed to serve media.",
			Code:    http.StatusInternalServerError,
		}
	}

	return nil
}
