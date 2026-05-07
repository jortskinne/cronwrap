package notifier

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/subtle-byte/cronwrap/internal/runner"
)

// LogNotifier writes notification events to an io.Writer (default: os.Stderr).
// It is useful for local debugging and audit trails without requiring an
// external service.
type LogNotifier struct {
	out    io.Writer
	prefix string
}

// NewLogNotifier returns a LogNotifier that writes to w.
// If w is nil, os.Stderr is used.
// prefix is an optional label prepended to every log line (e.g. the job name).
func NewLogNotifier(w io.Writer, prefix string) *LogNotifier {
	if w == nil {
		w = os.Stderr
	}
	return &LogNotifier{out: w, prefix: prefix}
}

// Notify writes a single log line describing the result.
func (l *LogNotifier) Notify(result runner.Result) error {
	status := "SUCCESS"
	if result.ExitCode != 0 {
		status = "FAILURE"
	}

	label := ""
	if l.prefix != "" {
		label = l.prefix + " "
	}

	_, err := fmt.Fprintf(
		l.out,
		"%s [cronwrap] %s%s exit=%d duration=%s\n",
		time.Now().UTC().Format(time.RFC3339),
		label,
		status,
		result.ExitCode,
		result.Duration.Round(time.Millisecond),
	)
	return err
}
