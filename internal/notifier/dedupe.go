package notifier

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/example/cronwrap/internal/runner"
)

// DedupeNotifier suppresses duplicate notifications for the same job+exit
// code combination within a configurable window.
type DedupeNotifier struct {
	inner  Notifier
	window time.Duration

	mu   sync.Mutex
	seen map[string]time.Time
}

// NewDedupeNotifier wraps inner and silences repeat notifications that share
// the same deduplication key within window.
func NewDedupeNotifier(inner Notifier, window time.Duration) *DedupeNotifier {
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &DedupeNotifier{
		inner:  inner,
		window: window,
		seen:   make(map[string]time.Time),
	}
}

// Notify forwards to the inner notifier only when the dedup key has not been
// seen within the configured window.
func (d *DedupeNotifier) Notify(result runner.Result) error {
	key := dedupeKey(result)
	now := time.Now()

	d.mu.Lock()
	last, exists := d.seen[key]
	if exists && now.Sub(last) < d.window {
		d.mu.Unlock()
		return nil
	}
	d.seen[key] = now
	d.mu.Unlock()

	return d.inner.Notify(result)
}

// dedupeKey returns a stable string that identifies a particular
// job-outcome combination.
func dedupeKey(r runner.Result) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s:%d", r.Command, r.ExitCode)
	return fmt.Sprintf("%x", h.Sum(nil))
}
