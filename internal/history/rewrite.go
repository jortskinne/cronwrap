package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Rewrite atomically replaces the store's file with the given records.
// It writes to a temporary file in the same directory and then renames it
// to ensure the operation is as atomic as the underlying filesystem allows.
func (s *Store) Rewrite(records []Record) error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("history rewrite: mkdir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".cronwrap-rewrite-*.jsonl")
	if err != nil {
		return fmt.Errorf("history rewrite: create temp: %w", err)
	}
	tmpName := tmp.Name()

	writeErr := func() error {
		enc := json.NewEncoder(tmp)
		for _, r := range records {
			if err := enc.Encode(r); err != nil {
				return fmt.Errorf("history rewrite: encode: %w", err)
			}
		}
		return tmp.Close()
	}()

	if writeErr != nil {
		os.Remove(tmpName)
		return writeErr
	}

	if err := os.Rename(tmpName, s.Path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("history rewrite: rename: %w", err)
	}

	return nil
}
