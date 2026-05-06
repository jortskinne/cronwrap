package notifier

import (
	"fmt"
	"time"
)

// RetryNotifier wraps a Notifier and retries on failure up to MaxAttempts times.
type RetryNotifier struct {
	inner       Notifier
	maxAttempts int
	delay       time.Duration
}

// NewRetryNotifier returns a Notifier that retries up to maxAttempts times
// with the given delay between attempts.
func NewRetryNotifier(n Notifier, maxAttempts int, delay time.Duration) Notifier {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &RetryNotifier{
		inner:       n,
		maxAttempts: maxAttempts,
		delay:       delay,
	}
}

// Notify attempts to notify up to maxAttempts times, returning nil on first
// success or the last error if all attempts fail.
func (r *RetryNotifier) Notify(result JobResult) error {
	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		if err := r.inner.Notify(result); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < r.maxAttempts && r.delay > 0 {
			time.Sleep(r.delay)
		}
	}
	return fmt.Errorf("notifier failed after %d attempt(s): %w", r.maxAttempts, lastErr)
}
