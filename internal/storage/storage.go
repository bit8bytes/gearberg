// Package storage defines the object storage contract used across the application.
package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"time"
)

// Byte is a float64 representing a number of bytes, used for human-readable formatting.
type Byte float64

// KB, MB, GB, TB are size constants in powers of 1024.
const (
	_       = iota // ignore first value by assigning to blank identifier
	KB Byte = 1 << (10 * iota)
	MB
	GB
	TB
)

func (b Byte) String() string {
	switch {
	case b >= TB:
		return fmt.Sprintf("%.2f TB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2f GB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2f MB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2f KB", b/KB)
	default:
		return fmt.Sprintf("%d B", int64(b))
	}
}

// ListEntry is the metadata returned by List. It intentionally omits the raw
// bytes — callers use URL to build a link rather than streaming content through
// the server.
type ListEntry struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// Options carries upload metadata. Size is required by S3's PutObject; omitting
// it forces the SDK to buffer the entire body in memory to compute the length.
type Options struct {
	Size        int64
	ContentType string
}

// driver is the internal interface implemented by each backend.
// put returns the Tink key ID embedded in the stored ciphertext, or 0 for
// unencrypted drivers (e.g. NoCipher).
type driver interface {
	put(ctx context.Context, key string, r io.Reader, opts Options) (uint32, error)
	get(ctx context.Context, key string) (io.ReadCloser, error)
	delete(ctx context.Context, key string) error
	list(ctx context.Context, prefix string) ([]ListEntry, error)
}

// NoCipher disables encryption. Data is stored in plaintext.
// Replace with a real cipher implementation before handling sensitive data.
type NoCipher struct{}

// Store is the public handle returned by Open. Callers define their own
// minimal interface for the methods they need (see database/sql pattern).
type Store struct {
	driver  driver
	Backend string
}

// Open returns a [*Store] for the given driver and DSN.
// Supported drivers: "local" (DSN: file:///path/to/dir or file://./relative).
// For the local driver, Open creates the base directory if it does not exist.
func Open(driverName, dsn string, logger *slog.Logger) (*Store, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: invalid dsn: %w", err)
	}

	switch driverName {
	case "local":
		// file://./uploads parses as host="." path="/uploads"; rejoin to get "./uploads".
		// file:///absolute parses as host="" path="/absolute"; u.Host is empty so dir=u.Path.
		dir := u.Host + u.Path
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("storage: create base dir %q: %w", dir, err)
		}
		return &Store{
			driver:  newLocal(dir, logger),
			Backend: "local",
		}, nil
	default:
		return nil, fmt.Errorf("storage: unsupported driver %q", driverName)
	}
}

// Put streams r directly to the backend. Returns the Tink key ID of the stored
// ciphertext, or 0 when no encryption key is tracked (e.g. NoCipher).
func (s *Store) Put(ctx context.Context, key string, r io.Reader, opts Options) (uint32, error) {
	return s.driver.put(ctx, key, r, opts)
}

// Get returns a ReadCloser for the object at key. The caller must close it.
func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.driver.get(ctx, key)
}

// Delete removes the object at key.
func (s *Store) Delete(ctx context.Context, key string) error {
	return s.driver.delete(ctx, key)
}

// List returns all objects whose key starts with prefix.
func (s *Store) List(ctx context.Context, prefix string) ([]ListEntry, error) {
	return s.driver.list(ctx, prefix)
}
