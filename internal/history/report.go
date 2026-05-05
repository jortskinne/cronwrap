package history

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// PrintReport writes a human-readable history report to w.
// It prints a summary header followed by the most recent records (up to limit).
// If limit <= 0 all records are printed.
func PrintReport(w io.Writer, records []Record, limit int) error {
	s := Summarize(records)

	fmt.Fprintf(w, "=== cronwrap history report ===\n")
	fmt.Fprintf(w, "Total runs : %d\n", s.Total)
	fmt.Fprintf(w, "Successes  : %d\n", s.Successes)
	fmt.Fprintf(w, "Failures   : %d\n", s.Failures)

	if s.Total > 0 {
		fmt.Fprintf(w, "Last run   : %s (%s)\n", s.LastRun.Format(time.RFC3339), s.LastStatus)
		fmt.Fprintf(w, "Avg runtime: %s\n", s.AvgRuntime.Round(time.Millisecond))
	}

	fmt.Fprintln(w)

	slice := records
	if limit > 0 && len(records) > limit {
		slice = records[len(records)-limit:]
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STARTED\tDURATION\tEXIT\tCOMMAND")
	for _, r := range slice {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			r.StartedAt.Format(time.RFC3339),
			r.Duration.Round(time.Millisecond),
			r.ExitCode,
			r.Command,
		)
	}
	return tw.Flush()
}
