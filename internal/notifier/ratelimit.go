package notifier

import (
	"fmt"
	"sync"
	"time"

	"github.com/droppedbars/cronwrap/internal/runner"
)

// RateLimitedNotifier wraps a Notifier and enforces a minimum interval
// between notifications for the same job, regardless of outcome.
type RateLimitedNotifier struct {
	inner    Notifier
	interval time.Duration
	mu       sync.Mutex
	lastSent map[string]time.Time
}

// NewRateLimitedNotifier returns a Notifier that suppresses notifications
// sent more frequently than interval for the same command.
func NewRateLimitedNotifier(inner Notifier, interval time.Duration) Notifier {
	return &RateLimitedNotifier{
		inner:    inner,
		interval: interval,
		lastSent: make(map[string]time.Time),
	}
}

func (r *RateLimitedNotifier) Notify(result runner.Result) error {
	if r.interval <= 0 {
		return r.inner.Notify(result)
	}

	key := rateLimitKey(result)

	r.mu.Lock()
	last, seen := r.lastSent[key]
	if seen && time.Since(last) < r.interval {
		r.mu.Unlock()
		return nil
	}
	r.lastSent[key] = time.Now()
	r.mu.Unlock()

	return r.inner.Notify(result)
}

func rateLimitKey(result runner.Result) string {
	return fmt.Sprintf("%s|%v", result.Command, result.Success)
}
