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
	"io"
	"net/http"

	"github.com/bit8bytes/gearberg/internal/httperr"
)

func (app *application) getMedia(w http.ResponseWriter, r *http.Request) *httperr.Error {
	id := r.PathValue("id")

	obj, err := app.services.storageManager.Get(r.Context(), id)
	if err != nil {
		return &httperr.Error{
			Error:   err,
			Message: "Media not found.",
			Code:    http.StatusNotFound,
		}
	}

	rc, err := app.services.storageManager.Open(r.Context(), id)
	if err != nil {
		return &httperr.Error{
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
		return &httperr.Error{
			Error:   err,
			Message: "Failed to serve media.",
			Code:    http.StatusInternalServerError,
		}
	}

	return nil
}
