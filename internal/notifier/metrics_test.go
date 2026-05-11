package notifier_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/exampleorg/cronwrap/internal/notifier"
	"github.com/exampleorg/cronwrap/internal/runner"
)

func metricsSuccessResult() runner.Result {
	return runner.Result{Command: "echo hi", ExitCode: 0, Stdout: "hi\n"}
}

func metricsFailResult() runner.Result {
	return runner.Result{Command: "false", ExitCode: 1, Stderr: "error"}
}

func TestMetricsCollector_RecordsSuccess(t *testing.T) {
	var buf bytes.Buffer
	mc := notifier.NewMetricsCollector(&buf)

	n := mc.Wrap(notifier.NotifierFunc(func(r runner.Result) error { return nil }))
	if err := n.Notify(metricsSuccessResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mc.Summary()
	out := buf.String()
	if !strings.Contains(out, "total=1") {
		t.Errorf("expected total=1 in output, got: %s", out)
	}
	if !strings.Contains(out, "success=1") {
		t.Errorf("expected success=1 in output, got: %s", out)
	}
	if !strings.Contains(out, "failures=0") {
		t.Errorf("expected failures=0 in output, got: %s", out)
	}
}

func TestMetricsCollector_RecordsFailure(t *testing.T) {
	var buf bytes.Buffer
	mc := notifier.NewMetricsCollector(&buf)

	n := mc.Wrap(notifier.NotifierFunc(func(r runner.Result) error {
		return errors.New("send failed")
	}))
	_ = n.Notify(metricsFailResult())

	mc.Summary()
	out := buf.String()
	if !strings.Contains(out, "failures=1") {
		t.Errorf("expected failures=1 in output, got: %s", out)
	}
	if !strings.Contains(out, "success=0") {
		t.Errorf("expected success=0 in output, got: %s", out)
	}
}

func TestMetricsCollector_MultipleNotifications(t *testing.T) {
	var buf bytes.Buffer
	mc := notifier.NewMetricsCollector(&buf)

	calls := 0
	n := mc.Wrap(notifier.NotifierFunc(func(r runner.Result) error {
		calls++
		if calls%2 == 0 {
			return errors.New("even call fails")
		}
		return nil
	}))

	for i := 0; i < 4; i++ {
		_ = n.Notify(metricsSuccessResult())
	}

	mc.Summary()
	out := buf.String()
	if !strings.Contains(out, "total=4") {
		t.Errorf("expected total=4, got: %s", out)
	}
	if !strings.Contains(out, "success=2") {
		t.Errorf("expected success=2, got: %s", out)
	}
	if !strings.Contains(out, "failures=2") {
		t.Errorf("expected failures=2, got: %s", out)
	}
}

func TestMetricsCollector_NilWriterDefaultsToStderr(t *testing.T) {
	mc := notifier.NewMetricsCollector(nil)
	n := mc.Wrap(notifier.NotifierFunc(func(r runner.Result) error { return nil }))
	if err := n.Notify(metricsSuccessResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// just ensure no panic; stderr output is acceptable
}

func TestMetricsCollector_LatencyRecorded(t *testing.T) {
	var buf bytes.Buffer
	mc := notifier.NewMetricsCollector(&buf)

	n := mc.Wrap(notifier.NotifierFunc(func(r runner.Result) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	}))
	_ = n.Notify(metricsSuccessResult())

	mc.Summary()
	out := buf.String()
	if !strings.Contains(out, "avg_latency=") {
		t.Errorf("expected avg_latency in output, got: %s", out)
	}
	// avg_latency should not be 0s
	if strings.Contains(out, "avg_latency=0s") {
		t.Errorf("expected non-zero avg_latency, got: %s", out)
	}
}
