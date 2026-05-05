package history

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"
)

func TestExportCSV_Empty(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	var buf bytes.Buffer
	if err := ExportCSV(store, &buf); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header only), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "id,") {
		t.Errorf("expected header line, got: %s", lines[0])
	}
}

func TestExportCSV_WithRecords(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	now := time.Now().UTC()
	records := []Record{
		{Job: "backup", StartedAt: now, DurationMs: 200, ExitCode: 0, Success: true, Output: "done"},
		{Job: "cleanup", StartedAt: now, DurationMs: 500, ExitCode: 1, Success: false, Output: "error occurred"},
	}
	for _, r := range records {
		if err := store.Append(r); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := ExportCSV(store, &buf); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2), got %d", len(rows))
	}
	if rows[1][1] != "backup" {
		t.Errorf("row 1 job = %q, want %q", rows[1][1], "backup")
	}
	if rows[2][5] != "false" {
		t.Errorf("row 2 success = %q, want %q", rows[2][5], "false")
	}
}

func TestExportCSV_SnippetTruncated(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	longOutput := strings.Repeat("x", 200)
	r := Record{Job: "big", StartedAt: time.Now().UTC(), DurationMs: 10, ExitCode: 0, Success: true, Output: longOutput}
	if err := store.Append(r); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var buf bytes.Buffer
	if err := ExportCSV(store, &buf); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	cr := csv.NewReader(&buf)
	rows, _ := cr.ReadAll()
	snippet := rows[1][6]
	if !strings.HasSuffix(snippet, "...") {
		t.Errorf("expected truncated snippet ending in '...', got: %s", snippet)
	}
	if len(snippet) > 130 {
		t.Errorf("snippet too long: %d chars", len(snippet))
	}
}
