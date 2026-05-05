package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// ExportCSV writes all records from the store to w in CSV format.
// Columns: id, job, started_at, duration_ms, exit_code, success, output_snippet.
func ExportCSV(store *Store, w io.Writer) error {
	records, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("export: read history: %w", err)
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{"id", "job", "started_at", "duration_ms", "exit_code", "success", "output_snippet"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("export: write header: %w", err)
	}

	for i, r := range records {
		snippet := r.Output
		if len(snippet) > 120 {
			snippet = snippet[:120] + "..."
		}
		row := []string{
			fmt.Sprintf("%d", i+1),
			r.Job,
			r.StartedAt.UTC().Format(time.RFC3339),
			fmt.Sprintf("%d", r.DurationMs),
			fmt.Sprintf("%d", r.ExitCode),
			fmt.Sprintf("%t", r.Success),
			snippet,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("export: write row %d: %w", i+1, err)
		}
	}

	if err := cw.Error(); err != nil {
		return fmt.Errorf("export: flush: %w", err)
	}
	return nil
}
