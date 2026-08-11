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
package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type localStorage struct {
	baseDir string
	logger  *slog.Logger
}

func newLocal(baseDir string, logger *slog.Logger) *localStorage {
	return &localStorage{baseDir: baseDir, logger: logger}
}

// safeKey rejects any key that would escape baseDir. Clean removes redundant
// separators and dot elements; IsLocal ensures the result stays within the
// tree (no absolute paths, no ".." components); Join then produces a safe
// absolute path under baseDir.
func safeKey(key string) bool {
	return filepath.IsLocal(filepath.Clean(filepath.FromSlash(key)))
}

func (s *localStorage) put(_ context.Context, key string, r io.Reader, opts Options) (uint32, error) {
	if !safeKey(key) {
		return 0, fmt.Errorf("storage: invalid path")
	}
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))
	tmpName, err := s.writeTmp(filepath.Dir(path), r, opts.Size)
	if err != nil {
		return 0, err
	}
	// Atomic rename: if this fails, clean up the dangling temp file.
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return 0, fmt.Errorf("storage: rename temp file: %w", err)
	}
	return 0, nil
}

// writeTmp writes r into a new temp file under dir and returns its path.
// On any error the temp file is removed before returning. On success the
// caller owns the file and must rename or remove it.
func (s *localStorage) writeTmp(dir string, r io.Reader, size int64) (string, error) {
	// Allow only the app to create directories and files to prevent security issues.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("storage: create dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return "", fmt.Errorf("storage: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	n, err := io.Copy(tmp, r)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("storage: copy: %w", err)
	}
	if size > 0 && n != size {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("storage: wrote %d bytes, expected %d", n, size)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return "", fmt.Errorf("storage: close temp file: %w", err)
	}
	return tmpName, nil
}

func (s *localStorage) get(_ context.Context, key string) (io.ReadCloser, error) {
	if !safeKey(key) {
		return nil, fmt.Errorf("storage: invalid path")
	}
	f, err := os.Open(filepath.Join(s.baseDir, filepath.FromSlash(key)))
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	return f, nil
}

func (s *localStorage) delete(_ context.Context, key string) error {
	if !safeKey(key) {
		return fmt.Errorf("storage: invalid path")
	}
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("storage: remove: %w", err)
	}
	s.cleanupAncestors(path)
	return nil
}

// cleanupAncestors removes empty parent directories up to (not including) baseDir.
// Stops at the first non-empty directory or any error; all errors are intentionally ignored.
func (s *localStorage) cleanupAncestors(childPath string) {
	for dir := filepath.Dir(childPath); dir != s.baseDir; dir = filepath.Dir(dir) {
		if os.Remove(dir) != nil {
			return
		}
	}
}

func (s *localStorage) list(_ context.Context, prefix string) ([]ListEntry, error) {
	if prefix != "" && !safeKey(prefix) {
		return nil, fmt.Errorf("storage: invalid path")
	}
	root := filepath.Join(s.baseDir, filepath.FromSlash(prefix))
	var objects []ListEntry
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("storage: stat %s: %w", path, err)
		}
		rel, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return fmt.Errorf("storage: rel path: %w", err)
		}
		objects = append(objects, ListEntry{
			Key:          filepath.ToSlash(rel),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("storage: walk %s: %w", root, err)
	}
	return objects, nil
}
