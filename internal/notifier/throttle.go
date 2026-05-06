package notifier

import (
	"fmt"
	"sync"
	"time"
)

// ThrottledNotifier wraps a Notifier and suppresses repeated notifications
// for the same job within a cooldown window. This prevents alert storms when
// a cron job fails on every run.
type ThrottledNotifier struct {
	inner    Notifier
	cooldown time.Duration
	mu       sync.Mutex
	lastSent map[string]time.Time
}

// NewThrottledNotifier wraps inner and suppresses duplicate notifications
// for the same job name within the given cooldown duration.
func NewThrottledNotifier(inner Notifier, cooldown time.Duration) *ThrottledNotifier {
	return &ThrottledNotifier{
		inner:    inner,
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
	}
}

// Notify forwards the notification to the inner notifier only if the cooldown
// period has elapsed since the last notification for the same job.
func (t *ThrottledNotifier) Notify(r JobResult) error {
	key := throttleKey(r)

	t.mu.Lock()
	last, seen := t.lastSent[key]
	if seen && time.Since(last) < t.cooldown {
		t.mu.Unlock()
		return nil
	}
	t.lastSent[key] = time.Now()
	t.mu.Unlock()

	return t.inner.Notify(r)
}

// Reset clears the throttle state for a specific job, allowing the next
// notification to be sent immediately regardless of the cooldown.
func (t *ThrottledNotifier) Reset(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastSent, jobName+":failure")
	delete(t.lastSent, jobName+":success")
}

func throttleKey(r JobResult) string {
	status := "success"
	if !r.Success {
		status = "failure"
	}
	return fmt.Sprintf("%s:%s", r.JobName, status)
}
