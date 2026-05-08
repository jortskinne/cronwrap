package notifier

import (
	"context"
	"sync"
	"time"

	"github.com/reugn/cronwrap/internal/runner"
)

// BatchNotifier accumulates results and flushes them as a single notification
// after a configurable window or when the buffer reaches a maximum size.
type BatchNotifier struct {
	inner    Notifier
	window   time.Duration
	maxSize  int
	mu       sync.Mutex
	buf      []runner.Result
	timer    *time.Timer
	stopOnce sync.Once
	stopCh   chan struct{}
}

// NewBatchNotifier returns a BatchNotifier that wraps inner.
// window is the maximum time to wait before flushing; maxSize flushes early
// when the buffer reaches that count. A maxSize of 0 disables size-based flushing.
func NewBatchNotifier(inner Notifier, window time.Duration, maxSize int) *BatchNotifier {
	b := &BatchNotifier{
		inner:   inner,
		window:  window,
		maxSize: maxSize,
		stopCh:  make(chan struct{}),
	}
	return b
}

// Notify buffers the result and flushes if the size limit is reached.
func (b *BatchNotifier) Notify(ctx context.Context, result runner.Result) error {
	b.mu.Lock()
	b.buf = append(b.buf, result)
	size := len(b.buf)

	if b.timer == nil && b.window > 0 {
		b.timer = time.AfterFunc(b.window, func() {
			b.mu.Lock()
			records := b.drain()
			b.mu.Unlock()
			if len(records) > 0 {
				_ = b.inner.Notify(ctx, mergeBatch(records))
			}
		})
	}

	var toFlush []runner.Result
	if b.maxSize > 0 && size >= b.maxSize {
		toFlush = b.drain()
		if b.timer != nil {
			b.timer.Stop()
			b.timer = nil
		}
	}
	b.mu.Unlock()

	if len(toFlush) > 0 {
		return b.inner.Notify(ctx, mergeBatch(toFlush))
	}
	return nil
}

// Flush forces an immediate send of any buffered results.
func (b *BatchNotifier) Flush(ctx context.Context) error {
	b.mu.Lock()
	records := b.drain()
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.mu.Unlock()

	if len(records) == 0 {
		return nil
	}
	return b.inner.Notify(ctx, mergeBatch(records))
}

// drain returns the current buffer and resets it. Caller must hold b.mu.
func (b *BatchNotifier) drain() []runner.Result {
	records := b.buf
	b.buf = nil
	return records
}

// mergeBatch combines multiple results into a single summary Result.
func mergeBatch(records []runner.Result) runner.Result {
	if len(records) == 1 {
		return records[0]
	}
	var combined runner.Result
	combined.Command = records[0].Command
	for i, r := range records {
		if r.ExitCode != 0 {
			combined.ExitCode = r.ExitCode
		}
		if i == 0 {
			combined.StartedAt = r.StartedAt
		}
		combined.Duration += r.Duration
		if r.Output != "" {
			if combined.Output != "" {
				combined.Output += "\n---\n"
			}
			combined.Output += r.Output
		}
	}
	return combined
}
