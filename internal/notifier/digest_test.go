package notifier

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cronwrap/internal/runner"
)

type captureNotifier struct {
	mu      sync.Mutex
	results []runner.Result
	err     error
}

func (c *captureNotifier) Notify(r runner.Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, r)
	return c.err
}

func (c *captureNotifier) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.results)
}

func (c *captureNotifier) last() runner.Result {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.results[len(c.results)-1]
}

func TestDigestNotifier_FlushSendsOnce(t *testing.T) {
	cap := &captureNotifier{}
	d := NewDigestNotifier(cap, 10*time.Second)

	_ = d.Notify(runner.Result{Command: "job1", ExitCode: 0, Duration: time.Second})
	_ = d.Notify(runner.Result{Command: "job2", ExitCode: 1, Duration: 2 * time.Second})
	_ = d.Notify(runner.Result{Command: "job3", ExitCode: 0, Duration: 500 * time.Millisecond})

	if err := d.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	if cap.count() != 1 {
		t.Fatalf("expected 1 notification, got %d", cap.count())
	}

	result := cap.last()
	if result.ExitCode == 0 {
		t.Errorf("expected non-zero exit code due to failure")
	}
	for _, want := range []string{"3 job(s)", "1 failure(s)", "job1", "job2", "job3"} {
		if !containsStr(result.Output, want) {
			t.Errorf("output missing %q; got:\n%s", want, result.Output)
		}
	}
}

func TestDigestNotifier_FlushEmpty(t *testing.T) {
	cap := &captureNotifier{}
	d := NewDigestNotifier(cap, 10*time.Second)

	if err := d.Flush(); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if cap.count() != 0 {
		t.Errorf("expected no notifications for empty flush")
	}
}

func TestDigestNotifier_WindowTrigger(t *testing.T) {
	cap := &captureNotifier{}
	d := NewDigestNotifier(cap, 50*time.Millisecond)

	_ = d.Notify(runner.Result{Command: "cron-task", ExitCode: 0, Duration: time.Millisecond})

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if cap.count() == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Error("digest was not flushed automatically after window elapsed")
}

func TestDigestNotifier_InnerError(t *testing.T) {
	cap := &captureNotifier{err: errors.New("send failed")}
	d := NewDigestNotifier(cap, 10*time.Second)

	_ = d.Notify(runner.Result{Command: "job", ExitCode: 0})

	if err := d.Flush(); err == nil {
		t.Error("expected error from inner notifier, got nil")
	}
}

func TestDigestNotifier_ZeroWindowUsesDefault(t *testing.T) {
	d := NewDigestNotifier(&captureNotifier{}, 0)
	if d.window != 5*time.Minute {
		t.Errorf("expected default window 5m, got %v", d.window)
	}
}
