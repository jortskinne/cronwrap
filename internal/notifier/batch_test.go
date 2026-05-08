package notifier_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/reugn/cronwrap/internal/notifier"
	"github.com/reugn/cronwrap/internal/runner"
)

type captureNotifier struct {
	mu      sync.Mutex
	calls   []runner.Result
	errToReturn error
}

func (c *captureNotifier) Notify(_ context.Context, r runner.Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, r)
	return c.errToReturn
}

func (c *captureNotifier) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.calls)
}

func (c *captureNotifier) last() runner.Result {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls[len(c.calls)-1]
}

func TestBatchNotifier_FlushOnMaxSize(t *testing.T) {
	cap := &captureNotifier{}
	b := notifier.NewBatchNotifier(cap, 10*time.Second, 3)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_ = b.Notify(ctx, runner.Result{Command: "echo", Output: "line"})
	}
	if cap.count() != 0 {
		t.Fatalf("expected 0 flushes before max, got %d", cap.count())
	}

	_ = b.Notify(ctx, runner.Result{Command: "echo", Output: "line3"})
	if cap.count() != 1 {
		t.Fatalf("expected 1 flush at max size, got %d", cap.count())
	}
}

func TestBatchNotifier_FlushOnWindow(t *testing.T) {
	cap := &captureNotifier{}
	b := notifier.NewBatchNotifier(cap, 50*time.Millisecond, 0)
	ctx := context.Background()

	_ = b.Notify(ctx, runner.Result{Command: "echo", Output: "hello"})
	if cap.count() != 0 {
		t.Fatal("should not flush immediately")
	}

	time.Sleep(120 * time.Millisecond)
	if cap.count() != 1 {
		t.Fatalf("expected 1 flush after window, got %d", cap.count())
	}
}

func TestBatchNotifier_ManualFlush(t *testing.T) {
	cap := &captureNotifier{}
	b := notifier.NewBatchNotifier(cap, 10*time.Second, 0)
	ctx := context.Background()

	_ = b.Notify(ctx, runner.Result{Command: "cmd1", Output: "a"})
	_ = b.Notify(ctx, runner.Result{Command: "cmd1", Output: "b"})

	if err := b.Flush(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.count() != 1 {
		t.Fatalf("expected 1 call after flush, got %d", cap.count())
	}
	result := cap.last()
	if result.Output == "" {
		t.Error("expected merged output")
	}
}

func TestBatchNotifier_FlushEmpty(t *testing.T) {
	cap := &captureNotifier{}
	b := notifier.NewBatchNotifier(cap, time.Second, 0)
	ctx := context.Background()

	if err := b.Flush(ctx); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if cap.count() != 0 {
		t.Fatal("expected no calls on empty flush")
	}
}

func TestBatchNotifier_SingleRecordPassthrough(t *testing.T) {
	cap := &captureNotifier{}
	b := notifier.NewBatchNotifier(cap, time.Second, 1)
	ctx := context.Background()

	want := runner.Result{Command: "ls", Output: "file.txt", ExitCode: 0}
	_ = b.Notify(ctx, want)

	if cap.count() != 1 {
		t.Fatalf("expected 1 flush, got %d", cap.count())
	}
	got := cap.last()
	if got.Command != want.Command || got.Output != want.Output {
		t.Errorf("result mismatch: got %+v, want %+v", got, want)
	}
}
