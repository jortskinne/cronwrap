package notifier

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/cronwrap/internal/runner"
)

type countingNotifier struct {
	calls atomic.Int32
	err   error
}

func (c *countingNotifier) Notify(_ runner.Result) error {
	c.calls.Add(1)
	return c.err
}

func failResult2() runner.Result {
	return runner.Result{Command: "backup.sh", ExitCode: 1}
}

func TestDedupeNotifier_FirstCallPasses(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner, time.Minute)

	if err := d.Notify(failResult2()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestDedupeNotifier_DuplicateSuppressed(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner, time.Minute)

	_ = d.Notify(failResult2())
	_ = d.Notify(failResult2())

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestDedupeNotifier_DifferentCommandsNotSuppressed(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner, time.Minute)

	r1 := runner.Result{Command: "job-a", ExitCode: 1}
	r2 := runner.Result{Command: "job-b", ExitCode: 1}

	_ = d.Notify(r1)
	_ = d.Notify(r2)

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", inner.calls.Load())
	}
}

func TestDedupeNotifier_WindowExpiry(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner, 50*time.Millisecond)

	_ = d.Notify(failResult2())
	time.Sleep(80 * time.Millisecond)
	_ = d.Notify(failResult2())

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls after window expiry, got %d", inner.calls.Load())
	}
}

func TestDedupeNotifier_PropagatesError(t *testing.T) {
	want := errors.New("send failed")
	inner := &countingNotifier{err: want}
	d := NewDedupeNotifier(inner, time.Minute)

	got := d.Notify(failResult2())
	if got != want {
		t.Fatalf("expected error %v, got %v", want, got)
	}
}

func TestDedupeNotifier_ZeroWindowUsesDefault(t *testing.T) {
	inner := &countingNotifier{}
	d := NewDedupeNotifier(inner, 0)

	if d.window != 5*time.Minute {
		t.Fatalf("expected default window 5m, got %v", d.window)
	}
}
