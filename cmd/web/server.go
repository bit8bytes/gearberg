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
