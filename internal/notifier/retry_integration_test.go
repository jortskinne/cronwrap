package notifier_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/notifier"
)

// TestRetryNotifier_WithMultiNotifier ensures RetryNotifier composes cleanly
// with MultiNotifier, retrying the whole multi-set on failure.
func TestRetryNotifier_WithMultiNotifier(t *testing.T) {
	a := &countingNotifier{failUntil: 1, err: errors.New("a fail")}
	b := &countingNotifier{failUntil: 0}

	multi := notifier.NewMulti(a, b)
	retried := notifier.NewRetryNotifier(multi, 3, 0)

	if err := retried.Notify(notifier.JobResult{JobName: "test"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	// a fails on attempt 1, succeeds on attempt 2
	if a.calls != 2 {
		t.Errorf("expected a.calls=2, got %d", a.calls)
	}
	// b always succeeds; multi stops on first error so b may not be called
	// on the failed attempt depending on multi ordering — just ensure no panic.
}

// TestRetryNotifier_ContextualErrorMessage checks the error wraps attempt count.
func TestRetryNotifier_ContextualErrorMessage(t *testing.T) {
	inner := &countingNotifier{failUntil: 99, err: errors.New("boom")}
	n := notifier.NewRetryNotifier(inner, 2, 0)
	err := n.Notify(notifier.JobResult{})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	for _, want := range []string{"2", "attempt"} {
		if !containsStr(msg, want) {
			t.Errorf("error message %q missing %q", msg, want)
		}
	}
}

// TestRetryNotifier_NoDelayOnLastAttempt ensures we don't sleep after the
// final failed attempt.
func TestRetryNotifier_NoDelayOnLastAttempt(t *testing.T) {
	inner := &countingNotifier{failUntil: 99, err: errors.New("fail")}
	delay := 50 * time.Millisecond
	n := notifier.NewRetryNotifier(inner, 2, delay)
	start := time.Now()
	_ = n.Notify(notifier.JobResult{})
	elapsed := time.Since(start)
	// Should sleep once (between attempt 1 and 2), not twice.
	if elapsed >= 2*delay {
		t.Errorf("slept too long: %v (expected < %v)", elapsed, 2*delay)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
