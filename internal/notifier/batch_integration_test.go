package notifier_test

import (
	"context"
	"testing"
	"time"

	"github.com/reugn/cronwrap/internal/notifier"
	"github.com/reugn/cronwrap/internal/runner"
)

func TestBatchNotifier_WithMultiNotifier(t *testing.T) {
	cap1 := &captureNotifier{}
	cap2 := &captureNotifier{}
	multi := notifier.NewMulti(cap1, cap2)
	batch := notifier.NewBatchNotifier(multi, time.Second, 2)
	ctx := context.Background()

	_ = batch.Notify(ctx, runner.Result{Command: "job", Output: "first"})
	_ = batch.Notify(ctx, runner.Result{Command: "job", Output: "second"})

	if cap1.count() != 1 {
		t.Errorf("cap1: expected 1 call, got %d", cap1.count())
	}
	if cap2.count() != 1 {
		t.Errorf("cap2: expected 1 call, got %d", cap2.count())
	}
}

func TestBatchNotifier_FailureExitCodePreserved(t *testing.T) {
	cap := &captureNotifier{}
	batch := notifier.NewBatchNotifier(cap, time.Second, 3)
	ctx := context.Background()

	_ = batch.Notify(ctx, runner.Result{Command: "job", ExitCode: 0, Output: "ok"})
	_ = batch.Notify(ctx, runner.Result{Command: "job", ExitCode: 1, Output: "fail"})
	_ = batch.Notify(ctx, runner.Result{Command: "job", ExitCode: 0, Output: "ok2"})

	if cap.count() != 1 {
		t.Fatalf("expected 1 flush, got %d", cap.count())
	}
	result := cap.last()
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code in merged result")
	}
}

func TestBatchNotifier_WindowAndFlushRace(t *testing.T) {
	cap := &captureNotifier{}
	batch := notifier.NewBatchNotifier(cap, 40*time.Millisecond, 0)
	ctx := context.Background()

	_ = batch.Notify(ctx, runner.Result{Command: "job", Output: "x"})
	// Manual flush before window fires.
	_ = batch.Flush(ctx)

	time.Sleep(80 * time.Millisecond)
	// Window timer should not double-deliver since buffer was drained.
	if cap.count() != 1 {
		t.Errorf("expected exactly 1 delivery, got %d", cap.count())
	}
}

func TestBatchNotifier_ConcurrentNotify(t *testing.T) {
	cap := &captureNotifier{}
	batch := notifier.NewBatchNotifier(cap, time.Second, 10)
	ctx := context.Background()

	doneCh := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			_ = batch.Notify(ctx, runner.Result{Command: "job", Output: "concurrent"})
			doneCh <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-doneCh
	}
	// All 10 goroutines together should trigger exactly 1 flush (maxSize=10).
	if cap.count() != 1 {
		t.Errorf("expected 1 flush from concurrent notifies, got %d", cap.count())
	}
}
