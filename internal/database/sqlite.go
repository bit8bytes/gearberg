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

//go:build !postgres

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"strings"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// Migrate applies all pending SQLite migrations from migrations to db.
func Migrate(ctx context.Context, db *sql.DB, migrations fs.FS) (int64, error) {
	return migrate(ctx, db, migrations, "sqlite")
}

// buildDSN validates dsn, converts it to a file: URI if needed, appends
// per-connection PRAGMA query params, and returns the expected journal mode.
// Only ":memory:" and "file:"-prefixed DSNs are accepted.
// The file: prefix is required so the driver parses the query string as a URI.
func buildDSN(dsn string) (fullDSN, journalMode string, err error) {
	if dsn != ":memory:" && !strings.HasPrefix(dsn, "file:") {
		return "", "", fmt.Errorf("buildDSN: sqlite DSN must be ':memory:' or start with 'file:', got %q", dsn)
	}

	base := dsn
	mode := "wal"
	if dsn == ":memory:" {
		base = "file::memory:"
		mode = "memory"
	}

	params := url.Values{}
	params.Add("_pragma", "journal_mode(WAL)")
	params.Add("_pragma", "foreign_keys(ON)")
	params.Add("_pragma", "busy_timeout(5000)")
	params.Add("_pragma", "cell_size_check(ON)")
	params.Add("_pragma", "trusted_schema(OFF)")
	params.Add("_pragma", "secure_delete(FAST)")
	params.Add("_pragma", "synchronous(NORMAL)")

	return base + "?" + params.Encode(), mode, nil
}

// New opens and returns an active SQLite connection pool for dsn.
// dsn must be ":memory:" or start with "file:" (e.g. "file:gearberg.db").
//
// Per-connection PRAGMAs are embedded in the URI DSN so the driver applies
// them on every new connection. Before the pool is returned, an integrity
// check, foreign-key check, and an initial optimize pass are run.
func New(ctx context.Context, dsn string) (*sql.DB, error) {
	fullDSN, expectedMode, err := buildDSN(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", fullDSN)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot open database at %q (use -db-dsn or DB_DSN): %w", dsn, err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(time.Hour)

	if err := integrityCheck(ctx, db); err != nil {
		return nil, err
	}

	if err := foreignKeyCheck(ctx, db); err != nil {
		return nil, err
	}

	if err := verifyJournalMode(ctx, expectedMode, db); err != nil {
		return nil, err
	}

	if err := optimize(ctx, db); err != nil {
		return nil, err
	}

	return db, nil
}

func foreignKeyCheck(ctx context.Context, db *sql.DB) (retErr error) {
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check;")
	if err != nil {
		return fmt.Errorf("foreign key query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("foreign key check close: %w", err)
		}
	}()

	hasViolation := rows.Next()
	rowsErr := rows.Err()

	if hasViolation {
		return ErrForeignKeyViolation
	}

	if rowsErr != nil {
		return fmt.Errorf("foreign key check: %w", rowsErr)
	}

	return nil
}

func integrityCheck(ctx context.Context, db *sql.DB) error {
	var integrity string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check;").Scan(&integrity); err != nil {
		return fmt.Errorf("integrity check query failed: %w", err)
	}

	if integrity != "ok" {
		return fmt.Errorf("integrity check failed: %s", integrity)
	}

	return nil
}

func verifyJournalMode(ctx context.Context, mode string, db *sql.DB) error {
	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		return fmt.Errorf("journal mode query failed: %w", err)
	}

	if journalMode != mode {
		return fmt.Errorf("expected journal_mode=wal, got %s", journalMode)
	}

	return nil
}

func optimize(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "PRAGMA optimize=0x10002;")
	if err != nil {
		return fmt.Errorf("could not optimize : %w", err)
	}
	return nil
}

// SessionStore returns an scs-compatible session store backed by db.
// Expired sessions are removed on the given cleanUpInterval.
func SessionStore(db *sql.DB, cleanUpInterval time.Duration) (scs.Store, error) {
	return sqlite3store.NewWithConfig(db, sqlite3store.Config{
		CleanUpInterval: cleanUpInterval,
	}), nil
}

// NormalizeError maps driver-specific SQLite errors to package-level sentinels.
// SQLITE_CONSTRAINT_UNIQUE returns [ErrUniqueConstraint];
// SQLITE_CONSTRAINT_PRIMARYKEY returns [ErrPrimaryKeyConstraint];
// SQLITE_CONSTRAINT_FOREIGNKEY returns [ErrForeignKeyViolation].
// All other errors are returned unchanged.
func NormalizeError(err error) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE:
			return ErrUniqueConstraint
		case sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY:
			return ErrPrimaryKeyConstraint
		case sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY:
			return ErrForeignKeyViolation
		case sqlite3.SQLITE_CONSTRAINT_TRIGGER:
			// SQLite enforces some FK constraints via triggers; in that case the
			// driver reports SQLITE_CONSTRAINT_TRIGGER instead of
			// SQLITE_CONSTRAINT_FOREIGNKEY.
			if strings.Contains(sqliteErr.Error(), "FOREIGN KEY") {
				return ErrForeignKeyViolation
			}
		}
	}
	return err
}
