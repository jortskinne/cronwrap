package notifier

import (
	"errors"
	"testing"
	"time"
)

type stubNotifier struct {
	called bool
	err    error
	last   JobResult
}

func (s *stubNotifier) Notify(r JobResult) error {
	s.called = true
	s.last = r
	return s.err
}

func TestMultiNotifier_AllCalled(t *testing.T) {
	a := &stubNotifier{}
	b := &stubNotifier{}
	m := NewMulti(a, b)

	r := JobResult{JobName: "test", Success: true, Duration: time.Second}
	if err := m.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called {
		t.Error("expected first notifier to be called")
	}
	if !b.called {
		t.Error("expected second notifier to be called")
	}
	if a.last.JobName != "test" {
		t.Errorf("unexpected job name: %s", a.last.JobName)
	}
}

func TestMultiNotifier_ReturnsFirstError(t *testing.T) {
	errA := errors.New("notifier A failed")
	a := &stubNotifier{err: errA}
	b := &stubNotifier{}
	m := NewMulti(a, b)

	err := m.Notify(JobResult{JobName: "job"})
	if err != errA {
		t.Errorf("expected errA, got %v", err)
	}
	// b must still have been called
	if !b.called {
		t.Error("expected second notifier to be called despite first error")
	}
}

func TestMultiNotifier_Empty(t *testing.T) {
	m := NewMulti()
	if err := m.Notify(JobResult{JobName: "noop"}); err != nil {
		t.Fatalf("unexpected error with no notifiers: %v", err)
	}
}

func TestJobResult_Fields(t *testing.T) {
	now := time.Now()
	r := JobResult{
		JobName:   "cleanup",
		Success:   false,
		ExitCode:  2,
		Output:    "stdout",
		Error:     "stderr",
		Duration:  3 * time.Second,
		StartedAt: now,
	}
	if r.ExitCode != 2 {
		t.Errorf("unexpected exit code: %d", r.ExitCode)
	}
	if r.StartedAt != now {
		t.Error("unexpected StartedAt")
	}
}
