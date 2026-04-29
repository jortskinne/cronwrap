// Package history provides run history tracking for cronwrap jobs.
package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Record represents a single job execution record.
type Record struct {
	JobName   string        `json:"job_name"`
	Command   string        `json:"command"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ns"`
	ExitCode  int           `json:"exit_code"`
	Success   bool          `json:"success"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
}

// Store manages persisting and retrieving run history records.
type Store struct {
	path string
}

// NewStore creates a new Store that writes records to the given file path.
func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

// Append adds a new record to the history file (one JSON object per line).
func (s *Store) Append(r Record) error {
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(r)
}

// ReadAll returns all records stored in the history file.
func (s *Store) ReadAll() ([]Record, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var records []Record
	dec := json.NewDecoder(f)
	for dec.More() {
		var r Record
		if err := dec.Decode(&r); err != nil {
			return records, err
		}
		records = append(records, r)
	}
	return records, nil
}
