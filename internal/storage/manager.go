// Package storage manages binary object storage across one or more named
// backends (e.g. "local", "s3"). Writes always go to the configured default
// backend. Reads are routed via the database, which records which backend each
// object was written to — so switching the default backend doesn't break access
// to existing objects.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	genstorage "github.com/bit8bytes/gearberg/database/queries/gen/storage"
	"github.com/segmentio/ksuid"
)

// ErrStorageQuotaExceeded is returned by CheckQuota when an upload would push
// the company over its configured storage limit.
var ErrStorageQuotaExceeded = errors.New("storage quota exceeded")

// Manager routes object storage operations across one or more backends.
// All writes go to the single currentBackend; reads are routed by consulting
// the database, which preserves backward compatibility when the backend changes.
type Manager struct {
	store           map[string]*Store
	currentBackend  string
	companyMaxBytes int64
	queries         *genstorage.Queries
}

// NewManager returns a Manager that routes writes to currentBackend and reads
// via the database-recorded backend for each object.
func NewManager(store map[string]*Store, currentBackend string, companyMaxBytes int64, db genstorage.DBTX) *Manager {
	return &Manager{
		store:           store,
		currentBackend:  currentBackend,
		companyMaxBytes: companyMaxBytes,
		queries:         genstorage.New(db),
	}
}

// Object is the domain model for a stored object.
type Object struct {
	ID              string
	CompanyID       string
	Key             string
	Backend         string
	Filename        string
	ContentType     string
	Size            int64
	CreatedAt       int64
	EncryptionKeyID int64
}

// Usage is a pre-computed snapshot of storage consumption for template rendering.
type Usage struct {
	BytesUsed int64
	BytesMax  int64
	Percent   int    // 0–100, clamped
	Used      string // e.g. "5.00 MB"
	Max       string // e.g. "5.00 GB"
}

// Info returns storage usage for the given company.
func (m *Manager) Info(ctx context.Context, companyID string) (Usage, error) {
	bytesUsed, err := m.queries.GetStorageUsedByCompany(ctx, companyID)
	if err != nil {
		return Usage{}, fmt.Errorf("storage.Info: %w", err)
	}
	var pct int
	if m.companyMaxBytes > 0 {
		pct = min(int(bytesUsed*100/m.companyMaxBytes), 100)
	}
	return Usage{
		BytesUsed: bytesUsed,
		BytesMax:  m.companyMaxBytes,
		Percent:   pct,
		Used:      Byte(bytesUsed).String(),
		Max:       Byte(m.companyMaxBytes).String(),
	}, nil
}

// CheckQuota returns ErrStorageQuotaExceeded if uploading size bytes would
// push the company over its quota. Call this before Put.
func (m *Manager) CheckQuota(ctx context.Context, companyID string, size int64) error {
	used, err := m.queries.GetStorageUsedByCompany(ctx, companyID)
	if err != nil {
		return fmt.Errorf("storage.CheckQuota: %w", err)
	}
	if used+size > m.companyMaxBytes {
		return ErrStorageQuotaExceeded
	}
	return nil
}

func storeFrom(store map[string]*Store, backend string) (*Store, error) {
	s, ok := store[backend]
	if !ok {
		return nil, fmt.Errorf("storage: unknown backend %q", backend)
	}
	return s, nil
}

// Put streams r to the backend and registers the object in the database.
func (m *Manager) Put(ctx context.Context, companyID, key, filename string, r io.Reader, opts Options) (*Object, error) {
	s, err := storeFrom(m.store, m.currentBackend)
	if err != nil {
		return nil, err
	}
	keyID, err := s.Put(ctx, key, r, opts)
	if err != nil {
		return nil, err
	}
	row, err := m.queries.Create(ctx, genstorage.CreateParams{
		ID:              ksuid.New().String(),
		CompanyID:       companyID,
		Key:             key,
		Backend:         s.Backend,
		Filename:        filename,
		ContentType:     opts.ContentType,
		Size:            opts.Size,
		EncryptionKeyID: int64(keyID),
	})
	if err != nil {
		_ = s.Delete(ctx, key)
		return nil, fmt.Errorf("storage.Put: register object: %w", err)
	}
	return recordFromRow(&row), nil
}

// Delete removes the object from the backend and the database.
func (m *Manager) Delete(ctx context.Context, id string) error {
	row, err := m.queries.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("storage.Delete: get record: %w", err)
	}
	s, err := storeFrom(m.store, row.Backend)
	if err != nil {
		return err
	}
	if err := s.Delete(ctx, row.Key); err != nil {
		return err
	}
	if err := m.queries.Delete(ctx, id); err != nil {
		return fmt.Errorf("storage.Delete: remove record: %w", err)
	}
	return nil
}

// Get returns the registry record for the given id.
func (m *Manager) Get(ctx context.Context, id string) (*Object, error) {
	row, err := m.queries.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("storage.Get: %w", err)
	}
	return recordFromRow(&row), nil
}

// Open returns a ReadCloser for the object. The caller must close it.
func (m *Manager) Open(ctx context.Context, id string) (io.ReadCloser, error) {
	row, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	s, err := storeFrom(m.store, row.Backend)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, row.Key)
}

// URL returns the public URL path for the given storage record ID.
func (m *Manager) URL(id string) string {
	return "/media/" + id
}

func recordFromRow(row *genstorage.StorageObject) *Object {
	return &Object{
		ID:              row.ID,
		CompanyID:       row.CompanyID,
		Key:             row.Key,
		Backend:         row.Backend,
		Filename:        row.Filename,
		ContentType:     row.ContentType,
		Size:            row.Size,
		CreatedAt:       row.CreatedAt,
		EncryptionKeyID: row.EncryptionKeyID,
	}
}
