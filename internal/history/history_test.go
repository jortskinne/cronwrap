package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

func TestStore_AppendAndReadAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	store, err := history.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	rec := history.Record{
		JobName:   "backup",
		Command:   "tar -czf /tmp/backup.tar.gz /data",
		StartedAt: time.Now().UTC().Truncate(time.Second),
		Duration:  2 * time.Second,
		ExitCode:  0,
		Success:   true,
		Output:    "done",
	}

	if err := store.Append(rec); err != nil {
		t.Fatalf("Append: %v", err)
	}

	records, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].JobName != rec.JobName {
		t.Errorf("JobName mismatch: got %q, want %q", records[0].JobName, rec.JobName)
	}
	if records[0].Success != rec.Success {
		t.Errorf("Success mismatch: got %v, want %v", records[0].Success, rec.Success)
	}
}

func TestStore_ReadAll_Empty(t *testing.T) {
	dir := t.TempDir()
	store, err := history.NewStore(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	records, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on missing file: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestStore_MultipleRecords(t *testing.T) {
	dir := t.TempDir()
	store, _ := history.NewStore(filepath.Join(dir, "history.jsonl"))

	for i := 0; i < 5; i++ {
		_ = store.Append(history.Record{JobName: "job", ExitCode: i})
	}

	records, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 5 {
		t.Errorf("expected 5 records, got %d", len(records))
	}
}

func TestStore_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nestedPath := filepath.Join(dir, "sub", "dir", "history.jsonl")
	_, err := history.NewStore(nestedPath)
	if err != nil {
		t.Fatalf("NewStore with nested path: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(nestedPath)); os.IsNotExist(err) {
		t.Error("expected directories to be created")
	}
}
