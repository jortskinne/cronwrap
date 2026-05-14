package notifier

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourorg/cronwrap/internal/runner"
)

// AuditEntry records a single notification attempt for audit purposes.
type AuditEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Command    string    `json:"command"`
	Success    bool      `json:"success"`
	ExitCode   int       `json:"exit_code"`
	Notifier   string    `json:"notifier"`
	Error      string    `json:"error,omitempty"`
	DurationMs int64     `json:"duration_ms"`
}

// AuditNotifier wraps another Notifier and writes an audit log entry for
// every notification attempt, regardless of outcome.
type AuditNotifier struct {
	inner    Notifier
	name     string
	writer   io.Writer
}

// NewAuditNotifier returns an AuditNotifier that decorates inner and writes
// JSON audit entries to w. If w is nil, os.Stderr is used. name identifies
// the wrapped notifier in the log.
func NewAuditNotifier(inner Notifier, name string, w io.Writer) *AuditNotifier {
	if w == nil {
		w = os.Stderr
	}
	return &AuditNotifier{inner: inner, name: name, writer: w}
}

// Notify delegates to the inner notifier and writes an audit entry.
func (a *AuditNotifier) Notify(result runner.Result) error {
	start := time.Now()
	err := a.inner.Notify(result)
	duration := time.Since(start)

	entry := AuditEntry{
		Timestamp:  start.UTC(),
		Command:    result.Command,
		Success:    result.Success,
		ExitCode:   result.ExitCode,
		Notifier:   a.name,
		DurationMs: duration.Milliseconds(),
	}
	if err != nil {
		entry.Error = err.Error()
	}

	data, jsonErr := json.Marshal(entry)
	if jsonErr == nil {
		fmt.Fprintf(a.writer, "%s\n", data)
	}

	return err
}
