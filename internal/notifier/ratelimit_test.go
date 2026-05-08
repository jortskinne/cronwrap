package notifier_test

import (
	"testing"
	"time"

	"github.com/droppedbars/cronwrap/internal/notifier"
	"github.com/droppedbars/cronwrap/internal/runner"
)

func makeRateLimitResult(cmd string, success bool) runner.Result {
	return runner.Result{Command: cmd, Success: success}
}

func TestRateLimitedNotifier_FirstCallPasses(t *testing.T) {
	called := 0
	mock := &mockNotifier{fn: func(r runner.Result) error { called++; return nil }}
	n := notifier.NewRateLimitedNotifier(mock, 10*time.Second)

	if err := n.Notify(makeRateLimitResult("backup.sh", false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Errorf("expected 1 call, got %d", called)
	}
}

func TestRateLimitedNotifier_SecondCallSuppressed(t *testing.T) {
	called := 0
	mock := &mockNotifier{fn: func(r runner.Result) error { called++; return nil }}
	n := notifier.NewRateLimitedNotifier(mock, 10*time.Second)

	res := makeRateLimitResult("backup.sh", false)
	_ = n.Notify(res)
	_ = n.Notify(res)

	if called != 1 {
		t.Errorf("expected 1 call after suppression, got %d", called)
	}
}

func TestRateLimitedNotifier_AllowsAfterInterval(t *testing.T) {
	called := 0
	mock := &mockNotifier{fn: func(r runner.Result) error { called++; return nil }}
	n := notifier.NewRateLimitedNotifier(mock, 10*time.Millisecond)

	res := makeRateLimitResult("backup.sh", false)
	_ = n.Notify(res)
	time.Sleep(20 * time.Millisecond)
	_ = n.Notify(res)

	if called != 2 {
		t.Errorf("expected 2 calls after interval elapsed, got %d", called)
	}
}

func TestRateLimitedNotifier_ZeroIntervalDisabled(t *testing.T) {
	called := 0
	mock := &mockNotifier{fn: func(r runner.Result) error { called++; return nil }}
	n := notifier.NewRateLimitedNotifier(mock, 0)

	res := makeRateLimitResult("backup.sh", false)
	_ = n.Notify(res)
	_ = n.Notify(res)
	_ = n.Notify(res)

	if called != 3 {
		t.Errorf("expected 3 calls with zero interval, got %d", called)
	}
}

func TestRateLimitedNotifier_DifferentCommandsIndependent(t *testing.T) {
	called := 0
	mock := &mockNotifier{fn: func(r runner.Result) error { called++; return nil }}
	n := notifier.NewRateLimitedNotifier(mock, 10*time.Second)

	_ = n.Notify(makeRateLimitResult("job-a", false))
	_ = n.Notify(makeRateLimitResult("job-b", false))

	if called != 2 {
		t.Errorf("expected 2 calls for different commands, got %d", called)
	}
}
