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
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

func (app *application) serve(ctx context.Context) error {
	return app.serveUnsecure(ctx)
}

func (app *application) newServer(ctx context.Context) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", app.options.Port),
		Handler:           http.TimeoutHandler(app.routes(), time.Second*20, ""),
		MaxHeaderBytes:    524_288,
		IdleTimeout:       time.Minute,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      15 * time.Second,
		ErrorLog:          slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}
}

func (app *application) serveUnsecure(ctx context.Context) error {
	srv := app.newServer(ctx)

	app.logger.Info("starting server", "addr", srv.Addr, "revision", revision)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		app.logger.Info("shutting down server...")
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
