package etagcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// File permissions for cached responses, which may contain private data.
const (
	fileStoreDirMode  = 0o700
	fileStoreFileMode = 0o600
)

// FileStore keeps one JSON file per entry in a directory, so a cache
// survives between runs of a command-line tool. Files are written with
// owner-only permissions and replaced atomically. Entries are never evicted;
// remove the directory to clear the cache.
type FileStore struct {
	dir string
}

// NewFileStore returns a FileStore in dir, creating it if needed.
func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, fileStoreDirMode); err != nil {
		return nil, fmt.Errorf("etagcache: create cache directory: %w", err)
	}
	return &FileStore{dir: dir}, nil
}

// Dir returns the cache directory.
func (s *FileStore) Dir() string { return s.dir }

func (s *FileStore) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:])+".json")
}

// Get implements Store.
func (s *FileStore) Get(key string) (*Entry, error) {
	data, err := os.ReadFile(s.path(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("etagcache: read entry: %w", err)
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		// A corrupt file is treated as a miss and overwritten by the next Set.
		return nil, nil
	}
	return &entry, nil
}

// Set implements Store.
func (s *FileStore) Set(key string, entry *Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("etagcache: encode entry: %w", err)
	}
	final := s.path(key)
	tmp, err := os.CreateTemp(s.dir, ".entry-*.tmp")
	if err != nil {
		return fmt.Errorf("etagcache: create temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()        // the write error is the one to report
		_ = os.Remove(tmpName) // best effort cleanup of a partial file
		return fmt.Errorf("etagcache: write entry: %w", err)
	}
	if err := tmp.Chmod(fileStoreFileMode); err != nil {
		_ = tmp.Close()        // the chmod error is the one to report
		_ = os.Remove(tmpName) // best effort cleanup
		return fmt.Errorf("etagcache: set entry permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName) // best effort cleanup
		return fmt.Errorf("etagcache: close entry: %w", err)
	}
	if err := os.Rename(tmpName, final); err != nil {
		_ = os.Remove(tmpName) // best effort cleanup
		return fmt.Errorf("etagcache: replace entry: %w", err)
	}
	return nil
}
