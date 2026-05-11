package notifier

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/exampleorg/cronwrap/internal/runner"
)

// MetricsCollector tracks notification attempt statistics across all notifiers.
type MetricsCollector struct {
	mu       sync.Mutex
	total    int
	success  int
	failures int
	suppressed int
	totalLatency time.Duration
	out      io.Writer
}

// NewMetricsCollector returns a MetricsCollector that writes summaries to w.
// If w is nil, os.Stderr is used.
func NewMetricsCollector(w io.Writer) *MetricsCollector {
	if w == nil {
		w = os.Stderr
	}
	return &MetricsCollector{out: w}
}

// Wrap returns a Notifier that records metrics around each call to inner.
func (m *MetricsCollector) Wrap(inner Notifier) Notifier {
	return &metricsNotifier{collector: m, inner: inner}
}

// Summary prints a human-readable metrics summary to the collector's writer.
func (m *MetricsCollector) Summary() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var avg time.Duration
	if m.total > 0 {
		avg = m.totalLatency / time.Duration(m.total)
	}
	fmt.Fprintf(m.out, "[metrics] total=%d success=%d failures=%d suppressed=%d avg_latency=%s\n",
		m.total, m.success, m.failures, m.suppressed, avg)
}

func (m *MetricsCollector) record(err error, suppressed bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.total++
	m.totalLatency += latency
	switch {
	case suppressed:
		m.suppressed++
	case err == nil:
		m.success++
	default:
		m.failures++
	}
}

type metricsNotifier struct {
	collector *MetricsCollector
	inner     Notifier
}

func (mn *metricsNotifier) Notify(result runner.Result) error {
	start := time.Now()
	err := mn.inner.Notify(result)
	latency := time.Since(start)
	mn.collector.record(err, false, latency)
	return err
}
