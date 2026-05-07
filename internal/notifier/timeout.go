package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/nzin/cronwrap/internal/runner"
)

// TimeoutNotifier wraps a Notifier and enforces a maximum duration for each
// Notify call. If the underlying notifier does not return within the deadline,
// the call is cancelled and an error is returned.
type TimeoutNotifier struct {
	inner   Notifier
	timeout time.Duration
}

// NewTimeoutNotifier returns a TimeoutNotifier that cancels calls to inner
// that exceed the given timeout. A timeout of zero disables enforcement and
// delegates directly to inner.
func NewTimeoutNotifier(inner Notifier, timeout time.Duration) *TimeoutNotifier {
	return &TimeoutNotifier{inner: inner, timeout: timeout}
}

// Notify calls the wrapped notifier with a deadline context. If the timeout
// elapses before the notifier returns, a descriptive error is returned.
func (t *TimeoutNotifier) Notify(ctx context.Context, result runner.Result) error {
	if t.timeout <= 0 {
		return t.inner.Notify(ctx, result)
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type outcome struct {
		err error
	}

	ch := make(chan outcome, 1)
	go func() {
		ch <- outcome{err: t.inner.Notify(ctx, result)}
	}()

	select {
	case out := <-ch:
		return out.err
	case <-ctx.Done():
		return fmt.Errorf("notifier timed out after %s: %w", t.timeout, ctx.Err())
	}
}
