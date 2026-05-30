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
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	htmlpkg "github.com/bit8bytes/gearberg/internal/html"
)

type application struct {
	logger   *slog.Logger
	options  *options
	html     *htmlpkg.HTML
	db       *sql.DB
	services *services
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	cmd, args := os.Args[1], os.Args[2:]
	switch cmd {
	case "serve":
		return runServe(args)
	case "verify":
		return runVerify(args)
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: gearberg <command> [flags]\n\nCommands:\n  serve   start the web server\n  verify  check configuration and database connectivity\n\nRun 'gearberg <command> -help' for command flags.\n")
}

func runServe(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	options, err := parseServeOptions(args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
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

	html := htmlpkg.New(log, cache, revision)

	db, err := setupDatabase(ctx, options)
	if err != nil {
		return fmt.Errorf("setup database: %w", err)
	}
	defer func(appDb *sql.DB) {
		if err := appDb.Close(); err != nil {
			log.InfoContext(ctx, "close database", "error", err.Error())
		}
	}(db)

	services, err := setupServices(db, options, log)
	if err != nil {
		return fmt.Errorf("setup services: %w", err)
	}

	app := &application{
		logger:   log,
		options:  options,
		html:     html,
		db:       db,
		services: services,
	}

	return app.serve(ctx)
}

func runVerify(args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	options, err := parseVerifyOptions(args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("parse options: %w", err)
	}

	log := setupLogger(options.LogLevel.Level())

	db, err := setupDatabase(ctx, options)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer func(appDb *sql.DB) {
		if err := appDb.Close(); err != nil {
			log.InfoContext(ctx, "close database", "error", err.Error())
		}
	}(db)

	log.InfoContext(ctx, "ok", "db_version", databaseVersion)
	return nil
}
