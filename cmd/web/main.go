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
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
)

type application struct {
	logger   *slog.Logger
	options  *options
	cache    map[string]*template.Template
	db       *sql.DB
	handlers http.Handler
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error running app: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	options, err := parseOptions()
	if err != nil {
		return fmt.Errorf("parse options: %w", err)
	}

	if options.Version {
		fmt.Printf("%s\n", revision)
		return nil
	}

	log := setupLogger(options.LogLevel.Level())

	cache, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	db, err := setupDatabase(ctx, options)
	if err != nil {
		return fmt.Errorf("setup database: %w", err)
	}
	defer func(appDb *sql.DB) {
		if err := appDb.Close(); err != nil {
			log.InfoContext(ctx, "close database", "error", err.Error())
		}
	}(db)

	handlers := setupHandlers(log, db, cache)

	app := &application{
		logger:   log,
		options:  options,
		cache:    cache,
		db:       db,
		handlers: handlers,
	}

	return app.serve(ctx)
}
