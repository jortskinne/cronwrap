package history

import (
	"os"
	"testing"
	"time"
)

var (
	now   = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	old   = now.Add(-48 * time.Hour)
	recent = now.Add(-1 * time.Hour)
)

func makeRetentionRecords() []Record {
	return []Record{
		{StartedAt: old, ExitCode: 0},
		{StartedAt: old.Add(time.Hour), ExitCode: 1},
		{StartedAt: recent, ExitCode: 0},
		{StartedAt: now, ExitCode: 0},
	}
}

func TestRetentionPolicy_NoOp(t *testing.T) {
	records := makeRetentionRecords()
	p := RetentionPolicy{}
	result := p.Apply(records)
	if len(result) != len(records) {
		t.Errorf("expected %d records, got %d", len(records), len(result))
	}
}

func TestRetentionPolicy_MaxAge(t *testing.T) {
	records := makeRetentionRecords()
	p := RetentionPolicy{MaxAge: 24 * time.Hour}
	result := p.Apply(records)
	for _, r := range result {
		if time.Since(r.StartedAt) > 24*time.Hour {
			t.Errorf("record older than MaxAge survived: %v", r.StartedAt)
		}
	}
}

func TestRetentionPolicy_MaxRecords(t *testing.T) {
	records := makeRetentionRecords()
	p := RetentionPolicy{MaxRecords: 2}
	result := p.Apply(records)
	if len(result) != 2 {
		t.Errorf("expected 2 records, got %d", len(result))
	}
	// Should keep the most recent
	if !result[0].StartedAt.Equal(recent) {
		t.Errorf("expected most recent records to be kept")
	}
}

func TestRetentionPolicy_Combined(t *testing.T) {
	records := makeRetentionRecords()
	p := RetentionPolicy{MaxAge: 24 * time.Hour, MaxRecords: 1}
	result := p.Apply(records)
	if len(result) != 1 {
		t.Errorf("expected 1 record, got %d", len(result))
	}
}

func TestApplyToStore(t *testing.T) {
	f, err := os.CreateTemp("", "retention-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	s := &Store{Path: f.Name()}
	for _, r := range makeRetentionRecords() {
		if err := s.Append(r); err != nil {
			t.Fatal(err)
		}
	}

	removed, err := ApplyToStore(s, RetentionPolicy{MaxRecords: 2})
	if err != nil {
		t.Fatalf("ApplyToStore: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	remaining, _ := s.ReadAll()
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining, got %d", len(remaining))
	}
}
