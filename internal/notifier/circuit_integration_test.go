package notifier_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fatih/cronwrap/internal/notifier"
	"github.com/fatih/cronwrap/internal/runner"
)

type countingNotifier struct {
	calls int
	errs  []error
}

func (c *countingNotifier) Notify(_ runner.Result) error {
	idx := c.calls
	c.calls++
	if idx < len(c.errs) {
		return c.errs[idx]
	}
	return nil
}

func TestCircuitBreaker_WithMultiNotifier(t *testing.T) {
	boom := errors.New("downstream unavailable")
	inner := &countingNotifier{errs: []error{boom, boom, boom, nil}}
	cb := notifier.NewCircuitBreaker(inner, 3, 10*time.Millisecond)

	result := runner.Result{ExitCode: 1, Stderr: "oops"}

	for i := 0; i < 3; i++ {
		_ = cb.Notify(result)
	}
	if cb.State() != notifier.CircuitOpen {
		t.Fatalf("expected circuit open after 3 failures")
	}

	// Immediate call should be blocked.
	err := cb.Notify(result)
	if err == nil {
		t.Fatal("expected blocked call to return error")
	}
	if inner.calls != 3 {
		t.Errorf("inner should have been called 3 times, got %d", inner.calls)
	}

	// Wait for reset window.
	time.Sleep(20 * time.Millisecond)

	// Probe — inner now returns nil.
	err = cb.Notify(result)
	if err != nil {
		t.Fatalf("probe should succeed, got: %v", err)
	}
	if cb.State() != notifier.CircuitClosed {
		t.Errorf("expected circuit closed after successful probe")
	}
}

func TestCircuitBreaker_ResetOnSuccess(t *testing.T) {
	inner := &countingNotifier{errs: []error{errors.New("e1"), errors.New("e2")}}
	cb := notifier.NewCircuitBreaker(inner, 5, time.Minute)

	_ = cb.Notify(runner.Result{ExitCode: 1})
	_ = cb.Notify(runner.Result{ExitCode: 1})

	if cb.State() != notifier.CircuitClosed {
		t.Errorf("circuit should still be closed below threshold")
	}

	// success resets counter
	if err := cb.Notify(runner.Result{ExitCode: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.State() != notifier.CircuitClosed {
		t.Error("expected circuit to remain closed after success")
	}
}
