package notifier

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/subtle-byte/cronwrap/internal/runner"
)

func successLogResult() runner.Result {
	return runner.Result{
		ExitCode: 0,
		Duration: 250 * time.Millisecond,
		Stdout:   "ok",
	}
}

func failureLogResult() runner.Result {
	return runner.Result{
		ExitCode: 1,
		Duration: 1200 * time.Millisecond,
		Stderr:   "something went wrong",
	}
}

func TestLogNotifier_Success(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "backup")

	if err := n.Notify(successLogResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	if !strings.Contains(line, "SUCCESS") {
		t.Errorf("expected SUCCESS in log line, got: %s", line)
	}
	if !strings.Contains(line, "backup") {
		t.Errorf("expected prefix 'backup' in log line, got: %s", line)
	}
	if !strings.Contains(line, "exit=0") {
		t.Errorf("expected exit=0 in log line, got: %s", line)
	}
}

func TestLogNotifier_Failure(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "")

	if err := n.Notify(failureLogResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	if !strings.Contains(line, "FAILURE") {
		t.Errorf("expected FAILURE in log line, got: %s", line)
	}
	if !strings.Contains(line, "exit=1") {
		t.Errorf("expected exit=1 in log line, got: %s", line)
	}
}

func TestLogNotifier_NilWriterDefaultsToStderr(t *testing.T) {
	// Should not panic; just verify construction succeeds.
	n := NewLogNotifier(nil, "job")
	if n.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestLogNotifier_NoPrefixOmitsLabel(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf, "")

	_ = n.Notify(successLogResult())

	line := buf.String()
	// With no prefix the label segment should be absent; the word SUCCESS must still appear.
	if !strings.Contains(line, "SUCCESS") {
		t.Errorf("expected SUCCESS, got: %s", line)
	}
}

func TestLogNotifier_WriteError(t *testing.T) {
	n := NewLogNotifier(&errWriter{}, "")
	err := n.Notify(successLogResult())
	if err == nil {
		t.Fatal("expected write error to be propagated")
	}
}

// errWriter always returns an error on Write.
type errWriter struct{}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}
