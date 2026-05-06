package notifier

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingNotifier struct {
	calls atomic.Int32
	err   error
}

func (c *countingNotifier) Notify(_ JobResult) error {
	c.calls.Add(1)
	return c.err
}

func failResult(name string) JobResult {
	return JobResult{JobName: name, Success: false, Output: "boom"}
}

func successResult(name string) JobResult {
	return JobResult{JobName: name, Success: true}
}

func TestThrottledNotifier_FirstCallPasses(t *testing.T) {
	inner := &countingNotifier{}
	tn := NewThrottledNotifier(inner, 5*time.Minute)

	if err := tn.Notify(failResult("job1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestThrottledNotifier_SecondCallSuppressed(t *testing.T) {
	inner := &countingNotifier{}
	tn := NewThrottledNotifier(inner, 5*time.Minute)

	_ = tn.Notify(failResult("job1"))
	_ = tn.Notify(failResult("job1"))

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call (throttled), got %d", inner.calls.Load())
	}
}

func TestThrottledNotifier_DifferentJobsNotThrottled(t *testing.T) {
	inner := &countingNotifier{}
	tn := NewThrottledNotifier(inner, 5*time.Minute)

	_ = tn.Notify(failResult("job1"))
	_ = tn.Notify(failResult("job2"))

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", inner.calls.Load())
	}
}

func TestThrottledNotifier_CooldownExpiry(t *testing.T) {
	inner := &countingNotifier{}
	tn := NewThrottledNotifier(inner, 10*time.Millisecond)

	_ = tn.Notify(failResult("job1"))
	time.Sleep(20 * time.Millisecond)
	_ = tn.Notify(failResult("job1"))

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls after cooldown, got %d", inner.calls.Load())
	}
}

func TestThrottledNotifier_Reset(t *testing.T) {
	inner := &countingNotifier{}
	tn := NewThrottledNotifier(inner, 5*time.Minute)

	_ = tn.Notify(failResult("job1"))
	tn.Reset("job1")
	_ = tn.Notify(failResult("job1"))

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", inner.calls.Load())
	}
}

func TestThrottledNotifier_SuccessAndFailureSeparateKeys(t *testing.T) {
	inner := &countingNotifier{}
	tn := NewThrottledNotifier(inner, 5*time.Minute)

	_ = tn.Notify(failResult("job1"))
	_ = tn.Notify(successResult("job1"))

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 calls (different status keys), got %d", inner.calls.Load())
	}
}

func TestThrottledNotifier_PropagatesError(t *testing.T) {
	expected := errors.New("send failed")
	inner := &countingNotifier{err: expected}
	tn := NewThrottledNotifier(inner, 5*time.Minute)

	err := tn.Notify(failResult("job1"))
	if !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}
