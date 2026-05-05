package history

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeRecords(n int) []Record {
	records := make([]Record, n)
	for i := 0; i < n; i++ {
		exitCode := 0
		if i%3 == 0 {
			exitCode = 1
		}
		records[i] = Record{
			Command:   "echo hello",
			StartedAt: time.Date(2024, 1, i+1, 12, 0, 0, 0, time.UTC),
			Duration:  time.Duration(i+1) * time.Second,
			ExitCode:  exitCode,
		}
	}
	return records
}

func TestPrintReport_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintReport(&buf, nil, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Total runs : 0") {
		t.Errorf("expected zero total in output, got:\n%s", buf.String())
	}
}

func TestPrintReport_ShowsHeader(t *testing.T) {
	records := makeRecords(3)
	var buf bytes.Buffer
	if err := PrintReport(&buf, records, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"STARTED", "DURATION", "EXIT", "COMMAND"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected column header %q in output", want)
		}
	}
}

func TestPrintReport_LimitApplied(t *testing.T) {
	records := makeRecords(10)
	var buf bytes.Buffer
	if err := PrintReport(&buf, records, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only last 3 records should appear; count occurrences of "echo hello"
	count := strings.Count(buf.String(), "echo hello")
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

func TestPrintReport_NoLimit(t *testing.T) {
	records := makeRecords(5)
	var buf bytes.Buffer
	if err := PrintReport(&buf, records, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := strings.Count(buf.String(), "echo hello")
	if count != 5 {
		t.Errorf("expected 5 rows, got %d", count)
	}
}

func TestPrintReport_TotalRunsCount(t *testing.T) {
	records := makeRecords(7)
	var buf bytes.Buffer
	if err := PrintReport(&buf, records, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The summary should always reflect the total number of records,
	// regardless of any display limit.
	if !strings.Contains(buf.String(), "Total runs : 7") {
		t.Errorf("expected 'Total runs : 7' in output, got:\n%s", buf.String())
	}
}
