// Package database provides shared error sentinels, migration helpers, and
// transaction utilities used by driver-specific implementations.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/pressly/goose/v3"
)

var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("record not found")
	// ErrUniqueConstraint is returned when an insert or update violates a unique index.
	ErrUniqueConstraint = errors.New("database has unique constraint violations")
	// ErrPrimaryKeyConstraint is returned when an insert violates a primary key constraint.
	ErrPrimaryKeyConstraint = errors.New("database has primary key constraint violations")
	// ErrForeignKeyViolation is returned when a write violates a foreign key reference.
	ErrForeignKeyViolation = errors.New("database has foreign key violations")
	// ErrLimitExceeded is returned when an operation exceeds a configured row or resource limit.
	ErrLimitExceeded = errors.New("entry limit exceeded")
)

// migrate applies all pending migrations from migrations to db and returns the
// resulting schema version. driver is used as both the goose dialect and the
// subdirectory name within the embedded FS (e.g. "sqlite" → migrations/sqlite/).
func migrate(ctx context.Context, db *sql.DB, migrations fs.FS, driver string) (version int64, err error) {
	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())

	if err = goose.SetDialect(driver); err != nil {
		return 0, fmt.Errorf("could not set goose dialect: %w", err)
	}

	if err = goose.UpContext(ctx, db, driver); err != nil {
		return 0, fmt.Errorf("goose migration failed: %w", err)
	}

	if version, err = goose.GetDBVersionContext(ctx, db); err != nil {
		return version, fmt.Errorf("could not get version: %w", err)
	}
	return
}

// NullInt64 converts a *int64 to sql.NullInt64.
func NullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// StringOrNil returns nil if s is empty, otherwise &s.
func StringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NullString converts a *string to sql.NullString.
func NullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

// Bool converts a bool to its SQLite integer representation (1 or 0).
func Bool(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Time converts a Unix timestamp to a UTC time.Time. Zero returns the zero value.
func Time(u int64) time.Time {
	if u == 0 {
		return time.Time{}
	}
	return time.Unix(u, 0).UTC()
}

// String returns the string value of s, or "" if not valid.
func String(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// StringPtr returns a pointer to the string value of s, or nil if not valid.
func StringPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	v := s.String
	return &v
}

// Int64 returns the int64 value of i, or 0 if not valid.
func Int64(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	return 0
}

// Int64Ptr returns a pointer to the int64 value of i, or nil if not valid.
func Int64Ptr(i sql.NullInt64) *int64 {
	if !i.Valid {
		return nil
	}
	v := i.Int64
	return &v
}

// NullFloat64 converts a float64 to sql.NullFloat64, treating 0 as NULL.
func NullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: v != 0}
}

// Float64 returns the float64 value of f, or 0 if not valid.
func Float64(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}
	return 0
}

// Float64Ptr returns a pointer to the float64 value of f, or nil if not valid.
func Float64Ptr(f sql.NullFloat64) *float64 {
	if !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}

// NullFloat64Ptr converts a *float64 to sql.NullFloat64.
func NullFloat64Ptr(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}
