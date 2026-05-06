package history

import (
	"os"
	"testing"
	"time"
)

func TestRewrite_ReplacesContent(t *testing.T) {
	f, err := os.CreateTemp("", "rewrite-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	s := &Store{Path: f.Name()}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 5; i++ {
		_ = s.Append(Record{StartedAt: base.Add(time.Duration(i) * time.Hour), ExitCode: 0})
	}

	keep := []Record{
		{StartedAt: base.Add(3 * time.Hour), ExitCode: 0},
		{StartedAt: base.Add(4 * time.Hour), ExitCode: 1},
	}

	if err := s.Rewrite(keep); err != nil {
		t.Fatalf("Rewrite: %v", err)
	}

	got, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[1].ExitCode != 1 {
		t.Errorf("expected last record exit code 1, got %d", got[1].ExitCode)
	}
}

func TestRewrite_EmptySlice(t *testing.T) {
	f, err := os.CreateTemp("", "rewrite-empty-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	s := &Store{Path: f.Name()}
	_ = s.Append(Record{StartedAt: time.Now(), ExitCode: 0})

	if err := s.Rewrite([]Record{}); err != nil {
		t.Fatalf("Rewrite empty: %v", err)
	}

	got, err := s.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 records after rewrite, got %d", len(got))
	}
}

func TestRewrite_CreatesParentDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "rewrite-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := dir + "/sub/history.jsonl"
	s := &Store{Path: path}

	records := []Record{{StartedAt: time.Now(), ExitCode: 0}}
	if err := s.Rewrite(records); err != nil {
		t.Fatalf("Rewrite with new dir: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}
