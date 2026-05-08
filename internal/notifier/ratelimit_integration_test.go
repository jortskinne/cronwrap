package notifier_test

import (
	"errors"
	"testing"
	"time"

	"github.com/droppedbars/cronwrap/internal/notifier"
	"github.com/droppedbars/cronwrap/internal/runner"
)

// TestRateLimitedNotifier_WithMultiNotifier verifies that rate limiting
// correctly wraps a MultiNotifier and suppresses redundant fan-out.
func TestRateLimitedNotifier_WithMultiNotifier(t *testing.T) {
	callsA, callsB := 0, 0
	a := &mockNotifier{fn: func(r runner.Result) error { callsA++; return nil }}
	b := &mockNotifier{fn: func(r runner.Result) error { callsB++; return nil }}
	multi := notifier.NewMulti(a, b)
	rl := notifier.NewRateLimitedNotifier(multi, 10*time.Second)

	res := runner.Result{Command: "sync.sh", Success: false}
	_ = rl.Notify(res)
	_ = rl.Notify(res) // suppressed

	if callsA != 1 || callsB != 1 {
		t.Errorf("expected each inner notifier called once, got a=%d b=%d", callsA, callsB)
	}
}

// TestRateLimitedNotifier_SuccessAndFailureSeparate verifies that success
// and failure outcomes are tracked independently for the same command.
func TestRateLimitedNotifier_SuccessAndFailureSeparate(t *testing.T) {
	called := 0
	mock := &mockNotifier{fn: func(r runner.Result) error { called++; return nil }}
	rl := notifier.NewRateLimitedNotifier(mock, 10*time.Second)

	_ = rl.Notify(runner.Result{Command: "job", Success: false})
	_ = rl.Notify(runner.Result{Command: "job", Success: true})

	if called != 2 {
		t.Errorf("expected 2 calls (different outcome keys), got %d", called)
	}
}

// TestRateLimitedNotifier_PropagatesError ensures errors from the inner
// notifier are returned to the caller.
func TestRateLimitedNotifier_PropagatesError(t *testing.T) {
	expected := errors.New("downstream failure")
	mock := &mockNotifier{fn: func(r runner.Result) error { return expected }}
	rl := notifier.NewRateLimitedNotifier(mock, 0) // zero = always pass through

	err := rl.Notify(runner.Result{Command: "job", Success: false})
	if !errors.Is(err, expected) {
		t.Errorf("expected propagated error, got %v", err)
	}
}
