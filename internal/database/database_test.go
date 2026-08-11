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
package database

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	"github.com/bit8bytes/gearberg/internal/database/migrations"
	_ "modernc.org/sqlite"
)

func TestNewSqlitePool_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := New(ctx, ":memory:")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	})

	if err := db.PingContext(ctx); err != nil {
		t.Errorf("PingContext() failed: %v", err)
	}
}

func TestNewSqlitePool_InvalidDSNFormat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := New(ctx, "/invalid/path/that/does/not/exist/test.db")
	if err == nil {
		t.Error("expected error for bare path DSN, got nil")
	}
}

func TestNewSqlitePool_InvalidPath(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := New(ctx, "file:/invalid/path/that/does/not/exist/test.db")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestMigrate_AppliesAllMigrations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := New(ctx, ":memory:")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	})

	version, err := Migrate(ctx, db, migrations.EmbedFS)
	if err != nil {
		t.Fatalf("Migrate() failed: %v", err)
	}

	if version <= 0 {
		t.Errorf("expected version > 0 after migration, got %d", version)
	}
}

func TestMigrate_IdempotentOnSecondRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := New(ctx, ":memory:")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	})

	v1, err := Migrate(ctx, db, migrations.EmbedFS)
	if err != nil {
		t.Fatalf("first Migrate() failed: %v", err)
	}

	v2, err := Migrate(ctx, db, migrations.EmbedFS)
	if err != nil {
		t.Fatalf("second Migrate() failed: %v", err)
	}

	if v1 != v2 {
		t.Errorf("expected same version on second run, got %d then %d", v1, v2)
	}
}

func TestMigrate_InvalidMigrationsFS(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := New(ctx, ":memory:")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	})

	emptyFS := fstest.MapFS{}
	_, err = Migrate(ctx, db, emptyFS)
	if err != nil {
		t.Logf("Migrate() with empty FS returned error (acceptable): %v", err)
	}
}
