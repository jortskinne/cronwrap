package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nzin/cronwrap/internal/runner"
)

// slowNotifier blocks for a configurable duration before returning.
type slowNotifier struct {
	delay time.Duration
	err   error
}

func (s *slowNotifier) Notify(ctx context.Context, _ runner.Result) error {
	select {
	case <-time.After(s.delay):
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestTimeoutNotifier_CompletesWithinDeadline(t *testing.T) {
	inner := &slowNotifier{delay: 10 * time.Millisecond}
	tn := NewTimeoutNotifier(inner, 500*time.Millisecond)

	if err := tn.Notify(context.Background(), runner.Result{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestTimeoutNotifier_ExceedsDeadline(t *testing.T) {
	inner := &slowNotifier{delay: 300 * time.Millisecond}
	tn := NewTimeoutNotifier(inner, 50*time.Millisecond)

	err := tn.Notify(context.Background(), runner.Result{})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded in error chain, got: %v", err)
	}
}

func TestTimeoutNotifier_ZeroTimeoutDisabled(t *testing.T) {
	// With zero timeout the inner notifier should be called directly.
	inner := &slowNotifier{delay: 20 * time.Millisecond, err: nil}
	tn := NewTimeoutNotifier(inner, 0)

	if err := tn.Notify(context.Background(), runner.Result{}); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestTimeoutNotifier_PropagatesInnerError(t *testing.T) {
	expected := errors.New("upstream failure")
	inner := &slowNotifier{delay: 5 * time.Millisecond, err: expected}
	tn := NewTimeoutNotifier(inner, 200*time.Millisecond)

	err := tn.Notify(context.Background(), runner.Result{})
	if !errors.Is(err, expected) {
		t.Errorf("expected wrapped inner error, got: %v", err)
	}
}

func TestTimeoutNotifier_ParentContextCancelled(t *testing.T) {
	inner := &slowNotifier{delay: 500 * time.Millisecond}
	tn := NewTimeoutNotifier(inner, 2*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := tn.Notify(ctx, runner.Result{})
	if err == nil {
		t.Fatal("expected error from cancelled parent context")
	}
}
