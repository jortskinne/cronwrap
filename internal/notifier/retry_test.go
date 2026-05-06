package notifier_test

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/notifier"
)

type countingNotifier struct {
	calls    int
	failUntil int
	err      error
}

func (c *countingNotifier) Notify(_ notifier.JobResult) error {
	c.calls++
	if c.calls <= c.failUntil {
		return c.err
	}
	return nil
}

func TestRetryNotifier_SuccessOnFirstAttempt(t *testing.T) {
	inner := &countingNotifier{failUntil: 0}
	n := notifier.NewRetryNotifier(inner, 3, 0)
	if err := n.Notify(notifier.JobResult{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestRetryNotifier_SuccessOnSecondAttempt(t *testing.T) {
	inner := &countingNotifier{failUntil: 1, err: errors.New("temporary")}
	n := notifier.NewRetryNotifier(inner, 3, 0)
	if err := n.Notify(notifier.JobResult{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if inner.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", inner.calls)
	}
}

func TestRetryNotifier_AllAttemptsFail(t *testing.T) {
	sentinel := errors.New("always fails")
	inner := &countingNotifier{failUntil: 99, err: sentinel}
	n := notifier.NewRetryNotifier(inner, 3, 0)
	err := n.Notify(notifier.JobResult{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped sentinel, got %v", err)
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetryNotifier_ZeroMaxAttemptsClamped(t *testing.T) {
	sentinel := errors.New("fail")
	inner := &countingNotifier{failUntil: 99, err: sentinel}
	n := notifier.NewRetryNotifier(inner, 0, 0)
	err := n.Notify(notifier.JobResult{})
	if err == nil {
		t.Fatal("expected error")
	}
	if inner.calls != 1 {
		t.Fatalf("expected exactly 1 call, got %d", inner.calls)
	}
}

func TestRetryNotifier_DelayBetweenAttempts(t *testing.T) {
	inner := &countingNotifier{failUntil: 1, err: errors.New("tmp")}
	delay := 10 * time.Millisecond
	n := notifier.NewRetryNotifier(inner, 2, delay)
	start := time.Now()
	_ = n.Notify(notifier.JobResult{})
	elapsed := time.Since(start)
	if elapsed < delay {
		t.Fatalf("expected at least %v delay, elapsed %v", delay, elapsed)
	}
}
