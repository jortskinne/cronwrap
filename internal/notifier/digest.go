package notifier

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cronwrap/internal/runner"
)

// DigestNotifier accumulates results over a window and sends a single
// summary notification instead of one per job execution.
type DigestNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	window   time.Duration
	records  []runner.Result
	timer    *time.Timer
}

// NewDigestNotifier returns a DigestNotifier that batches results and
// flushes a summary to inner after window duration of inactivity.
func NewDigestNotifier(inner Notifier, window time.Duration) *DigestNotifier {
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &DigestNotifier{
		inner:  inner,
		window: window,
	}
}

// Notify queues the result and resets the flush timer.
func (d *DigestNotifier) Notify(result runner.Result) error {
	d.mu.Lock()
	d.records = append(d.records, result)
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.window, func() {
		_ = d.Flush()
	})
	d.mu.Unlock()
	return nil
}

// Flush immediately sends a digest of all queued results and clears the queue.
func (d *DigestNotifier) Flush() error {
	d.mu.Lock()
	records := d.records
	d.records = nil
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.mu.Unlock()

	if len(records) == 0 {
		return nil
	}

	summary := buildDigestResult(records)
	return d.inner.Notify(summary)
}

// buildDigestResult merges multiple results into a single summary Result.
func buildDigestResult(records []runner.Result) runner.Result {
	var (
		failures  int
		lines     []string
		exitCode  int
		lastCmd   string
	)

	for _, r := range records {
		status := "OK"
		if r.ExitCode != 0 {
			failures++
			status = fmt.Sprintf("FAIL(%d)", r.ExitCode)
			if exitCode == 0 {
				exitCode = r.ExitCode
			}
		}
		lines = append(lines, fmt.Sprintf("[%s] %s — %s", status, r.Command, r.Duration.Round(time.Millisecond)))
		lastCmd = r.Command
	}

	header := fmt.Sprintf("Digest: %d job(s), %d failure(s)\n", len(records), failures)
	return runner.Result{
		Command:  lastCmd,
		Output:   header + strings.Join(lines, "\n"),
		ExitCode: exitCode,
		Duration: 0,
	}
}
