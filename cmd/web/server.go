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

// serve selects the appropriate server implementation based on the TLS mode.
// It runs either: An autocert-enabled server, a local server with manually provisioned TLS certificates, or in unsecure http mode.
func (app *application) serve(ctx context.Context) error {
	if app.options.TLS.Mode == "local" {
		return app.serveCertAndKey(ctx)
	}
	return app.serveUnsecure(ctx)
}

func (app *application) newServer(ctx context.Context) (*http.Server, error) {
	routes, err := app.routes()
	if err != nil {
		return nil, fmt.Errorf("newServer: %w", err)
	}
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", app.options.Port),
		Handler:           http.TimeoutHandler(routes, time.Second*20, ""),
		MaxHeaderBytes:    524_288,
		IdleTimeout:       time.Minute,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      15 * time.Second,
		ErrorLog:          slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}, nil
}

func (app *application) serveUnsecure(ctx context.Context) error {
	srv, err := app.newServer(ctx)
	if err != nil {
		return err
	}

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

// serveCertAndKey is used in deployment environments where TLS certificates
// are provisioned manually and provided via the Cert and Key config fields.
func (app *application) serveCertAndKey(ctx context.Context) error {
	srv, err := app.newServer(ctx)
	if err != nil {
		return err
	}

	app.logger.Info("starting server",
		"addr", srv.Addr,
		"revision", revision,
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.ListenAndServeTLS(app.options.TLS.CertPath, app.options.TLS.KeyPath); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve tls: %w", err)
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
		return fmt.Errorf("serveCertAndKey: %w", err)
	}
	return nil
}
