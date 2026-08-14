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
	"fmt"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	htmlpkg "github.com/bit8bytes/gearberg/internal/html"
)

type application struct {
	logger  *slog.Logger
	options *options
	html    *htmlpkg.HTML
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
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

	base, cache, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("load templates: %w", err)
	}

	app := &application{
		logger:  log,
		options: options,
		html:    htmlpkg.New(log, base, cache, revision),
	}

	return app.serve(ctx)
}
