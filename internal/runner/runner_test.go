package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/runner"
)

func TestRun_Success(t *testing.T) {
	result := runner.Run(context.Background(), "echo", []string{"hello"}, 0)

	if !result.Success() {
		t.Fatalf("expected success, got exit code %d, err: %v", result.ExitCode, result.Err)
	}
	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got: %q", result.Stdout)
	}
	if result.Duration <= 0 {
		t.Errorf("expected positive duration, got %v", result.Duration)
	}
}

func TestRun_Failure(t *testing.T) {
	result := runner.Run(context.Background(), "false", nil, 0)

	if result.Success() {
		t.Fatal("expected failure, got success")
	}
	if result.ExitCode == 0 {
		t.Errorf("expected non-zero exit code, got 0")
	}
}

func TestRun_StderrCapture(t *testing.T) {
	result := runner.Run(context.Background(), "sh", []string{"-c", "echo errout >&2"}, 0)

	if !strings.Contains(result.Stderr, "errout") {
		t.Errorf("expected stderr to contain 'errout', got: %q", result.Stderr)
	}
}

func TestRun_Timeout(t *testing.T) {
	start := time.Now()
	result := runner.Run(context.Background(), "sleep", []string{"10"}, 100*time.Millisecond)
	elapsed := time.Since(start)

	if result.Success() {
		t.Fatal("expected timeout failure, got success")
	}
	if elapsed > 2*time.Second {
		t.Errorf("command took too long despite timeout: %v", elapsed)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	result := runner.Run(context.Background(), "nonexistent_command_xyz", nil, 0)

	if result.Success() {
		t.Fatal("expected failure for invalid command")
	}
	if result.ExitCode != -1 {
		t.Errorf("expected exit code -1 for missing binary, got %d", result.ExitCode)
	}
}
