package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/runner"
)

func successAuditResult() runner.Result {
	return runner.Result{
		Command:   "echo hello",
		Success:   true,
		ExitCode:  0,
		StartTime: time.Now(),
		Duration:  10 * 1000000,
	}
}

func failAuditResult() runner.Result {
	return runner.Result{
		Command:   "false",
		Success:   false,
		ExitCode:  1,
		StartTime: time.Now(),
		Duration:  5 * 1000000,
	}
}

func TestAuditNotifier_WritesEntryOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	inner := &mockNotifier{}
	an := NewAuditNotifier(inner, "slack", &buf)

	if err := an.Notify(successAuditResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry AuditEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse audit entry: %v", err)
	}
	if entry.Command != "echo hello" {
		t.Errorf("expected command 'echo hello', got %q", entry.Command)
	}
	if !entry.Success {
		t.Error("expected success=true")
	}
	if entry.Notifier != "slack" {
		t.Errorf("expected notifier 'slack', got %q", entry.Notifier)
	}
	if entry.Error != "" {
		t.Errorf("expected no error field, got %q", entry.Error)
	}
}

func TestAuditNotifier_WritesErrorOnFailure(t *testing.T) {
	var buf bytes.Buffer
	inner := &mockNotifier{err: errors.New("send failed")}
	an := NewAuditNotifier(inner, "email", &buf)

	err := an.Notify(failAuditResult())
	if err == nil {
		t.Fatal("expected error from inner notifier")
	}

	var entry AuditEntry
	if jsonErr := json.Unmarshal(buf.Bytes(), &entry); jsonErr != nil {
		t.Fatalf("failed to parse audit entry: %v", jsonErr)
	}
	if entry.Error != "send failed" {
		t.Errorf("expected error 'send failed', got %q", entry.Error)
	}
}

func TestAuditNotifier_NilWriterDefaultsToStderr(t *testing.T) {
	// Should not panic when w is nil.
	inner := &mockNotifier{}
	an := NewAuditNotifier(inner, "webhook", nil)
	if an.writer == nil {
		t.Error("writer should default to os.Stderr, not nil")
	}
}

func TestAuditNotifier_EntryContainsDuration(t *testing.T) {
	var buf bytes.Buffer
	inner := &mockNotifier{}
	an := NewAuditNotifier(inner, "discord", &buf)

	_ = an.Notify(successAuditResult())

	if !strings.Contains(buf.String(), "duration_ms") {
		t.Error("audit entry should contain duration_ms field")
	}
}
