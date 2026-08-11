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
	"log/slog"
	"os"
	"testing"

	"github.com/bit8bytes/gearberg/internal/accounts"
	htmlpkg "github.com/bit8bytes/gearberg/internal/html"
	mailerpkg "github.com/bit8bytes/gearberg/pkg/mailer"
)

var (
	ts    *testServer
	app   *application
	alice *testUser
	bob   *testUser
)

func TestMain(m *testing.M) {
	code, err := runTests(m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(code)
}

// runTests exists so deferred cleanup runs before os.Exit.
func runTests(m *testing.M) (int, error) {
	var cleanup func()
	var err error
	app, ts, cleanup, err = startTestServer()
	if err != nil {
		return 1, err
	}
	defer cleanup()

	ctx := context.Background()
	if err := seedTestUsers(ctx); err != nil {
		return 1, err
	}

	return m.Run(), nil
}

// seedTestUsers creates alice and bob directly via the service layer so tests
// never need to sign up through HTTP. This avoids shared-session races where
// withGuest silently blocks a signup POST when a user is already authenticated.
func seedTestUsers(ctx context.Context) error {
	var err error
	alice, err = seedTestUser(ctx, aliceEmail, passwd1)
	if err != nil {
		return fmt.Errorf("seed alice: %w", err)
	}
	bob, err = seedTestUser(ctx, bobEmail, passwd2)
	if err != nil {
		return fmt.Errorf("seed bob: %w", err)
	}
	return nil
}

func seedTestUser(ctx context.Context, email, password string) (*testUser, error) {
	_, orgID, err := app.services.accounts.SignUp(ctx, accounts.SignUpParams{
		Email:          email,
		Password:       password,
		RepeatPassword: password,
	})
	if err != nil {
		return nil, fmt.Errorf("seedTestUser: %w", err)
	}
	return &testUser{
		email:    email,
		password: password,
		orgID:    orgID,
		ts:       ts,
	}, nil
}

// startTestServer mirrors runServe() in main.go: wires up the full app and HTTP
// test server from scratch using the same setup helpers. The returned cleanup
// func closes the DB and removes the temp directory; call it (via defer) before
// os.Exit.
func startTestServer() (*application, *testServer, func(), error) {
	dir, err := os.MkdirTemp("", "testmain-*")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("startTestServer: temp dir: %w", err)
	}

	opts := testOptions(dir)
	log := setupLogger(slog.LevelError)

	base, cache, err := parseTemplates()
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, nil, nil, fmt.Errorf("startTestServer: parse templates: %w", err)
	}

	html := htmlpkg.New(log, base, cache, revision)

	ctx := context.Background()

	db, err := setupDatabase(ctx, opts)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, nil, nil, fmt.Errorf("startTestServer: setup database: %w", err)
	}

	if err := seedReferenceData(ctx, db); err != nil {
		_ = db.Close()
		_ = os.RemoveAll(dir)
		return nil, nil, nil, fmt.Errorf("startTestServer: seed reference data: %w", err)
	}

	scsMgr, err := setupSCS(db)
	if err != nil {
		_ = db.Close()
		_ = os.RemoveAll(dir)
		return nil, nil, nil, fmt.Errorf("startTestServer: setup scs: %w", err)
	}
	sessionMgr := setupSessionManager(scsMgr)

	services, err := setupServices(db, opts, log, mailerpkg.LogMailer{})
	if err != nil {
		_ = db.Close()
		_ = os.RemoveAll(dir)
		return nil, nil, nil, fmt.Errorf("startTestServer: setup services: %w", err)
	}

	application := &application{
		logger:   log,
		options:  opts,
		html:     html,
		db:       db,
		session:  sessionMgr,
		services: services,
	}

	handler, err := application.routes()
	if err != nil {
		_ = db.Close()
		_ = os.RemoveAll(dir)
		return nil, nil, nil, fmt.Errorf("startTestServer: routes: %w", err)
	}

	srv := newTestServer(handler)

	cleanup := func() {
		srv.Close()
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}

	return application, srv, cleanup, nil
}
