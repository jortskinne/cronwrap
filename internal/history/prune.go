package history

import (
	"encoding/json"
	"os"
	"sort"
)

// Prune removes old records from the history file, keeping only the most
// recent maxRecords entries. If maxRecords is zero or negative, Prune is
// a no-op. The file is rewritten atomically using a temporary file.
func (s *Store) Prune(maxRecords int) error {
	if maxRecords <= 0 {
		return nil
	}

	records, err := s.ReadAll()
	if err != nil {
		return err
	}

	if len(records) <= maxRecords {
		return nil
	}

	// Sort ascending by StartedAt so we can keep the newest.
	sort.Slice(records, func(i, j int) bool {
		return records[i].StartedAt.Before(records[j].StartedAt)
	})

	records = records[len(records)-maxRecords:]

	tmpPath := s.path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			f.Close()
			os.Remove(tmpPath)
			return err
		}
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, s.path)
}
