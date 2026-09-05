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

	"github.com/bit8bytes/gearberg/internal/flash"
	htmlpkg "github.com/bit8bytes/gearberg/internal/html"
	"github.com/bit8bytes/gearberg/internal/templates"
	"github.com/bit8bytes/gearberg/internal/templates/pages"
)

// application is the dependency container for the web server. Fields are
// initialized once in runServe and shared across all HTTP handlers via method
// receivers; the alternative of passing each dependency as a function argument
// does not scale past a handful of handlers.
type application struct {
	logger *slog.Logger

	options *options

	html *htmlpkg.HTML

	// db is the raw connection pool. Handlers must not use it directly;
	// all queries go through services so business rules are enforced in one place.
	db *sql.DB

	session sessionManager

	services *services

	flash flash.Manager

	// oidcProviders is keyed by provider name. Providers are configured at
	// startup because OIDC discovery is a network call that should not happen per-request.
	oidcProviders map[string]*oidcProvider
}

// main entry point of the app.
func main() {
	// run is called to allow calling defer inside run() to gracefully stop the app.
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	// The first argument is the command, the rest are flags for that command.
	// This allows us to have different flags for different commands.
	cmd, args := os.Args[1], os.Args[2:]
	switch cmd {

	case "serve":
		// serve starts the server and listens on the configured port via options
		return runServe(args)
		// check verifies the provided configuration and returns nothing if successful.
	case "check":
		return runCheck(args)
	case "version":
		fmt.Printf("%s\n", revision)
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: gearberg <command> [flags]\n\nCommands:\n  serve   start the web server\n  check   check configuration and database connectivity\n\nRun 'gearberg <command> -help' for command flags.\n")
}

// runServe glues together the internal and external packages to start the web server.
func runServe(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	options, err := parseOptions(args)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("parse options: %w", err)
	}

	log := setupLogger(options.LogLevel.Level())

	templates := templates.New(templates.WithFuncs(templateFuncs()))
	base, cache, err := templates.Parse(pages.All)
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	db, err := setupDatabase(ctx, options)
	if err != nil {
		return fmt.Errorf("setup database: %w", err)
	}
	defer func(appDB *sql.DB) {
		_ = appDB.Close()
	}(db)

	if err := seedReferenceData(ctx, db); err != nil {
		return fmt.Errorf("seed reference data: %w", err)
	}

	scsMgr, err := setupSCS(db)
	if err != nil {
		return fmt.Errorf("setup session manager: %w", err)
	}
	sessionManager := setupSessionManager(scsMgr)
	flashManager := flash.NewSessionManager(sessionManager)

	html := htmlpkg.New(log, base, cache, revision, htmlpkg.WithFlashManager(flashManager))

	mailer := setupMailer(options, log)

	services, err := setupServices(db, options, log, mailer)
	if err != nil {
		return fmt.Errorf("setup services: %w", err)
	}

	oidcCtx, oidcCancel := context.WithTimeout(ctx, 10*time.Second)
	defer oidcCancel()

	oidcProviders, err := setupOIDCProviders(oidcCtx, options.OIDCProviders, options.BaseURL)
	if err != nil {
		return fmt.Errorf("setup oidc providers: %w", err)
	}

	// glue together the dependencies
	app := &application{
		logger:        log,
		options:       options,
		html:          html,
		db:            db,
		session:       sessionManager,
		flash:         flashManager,
		services:      services,
		oidcProviders: oidcProviders,
	}

	fmt.Printf("\033]8;;https://gearberg.org\033\\Gearberg\033]8;;\033\\ is running now. Open \033]8;;%s\033\\%s\033]8;;\033\\\n", options.BaseURL, options.BaseURL)

	return app.serve(ctx)
}

func runCheck(args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	options, err := parseOptions(args)
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
		_ = appDb.Close()
	}(db)

	if err := seedReferenceData(ctx, db); err != nil {
		return fmt.Errorf("seed reference data: %w", err)
	}
	log.InfoContext(ctx, "database ok", "db_version", databaseVersion)

	if len(options.OIDCProviders) == 0 {
		log.InfoContext(ctx, "oidc providers not configured")
	}
	for name, p := range options.OIDCProviders {
		if _, err := setupOIDCProvider(ctx, p, ""); err != nil {
			return fmt.Errorf("oidc-provider %s: %w", name, err)
		}
		log.InfoContext(ctx, "oidc provider ok", "name", name, "issuer", p.IssuerURL)
	}

	return nil
}
