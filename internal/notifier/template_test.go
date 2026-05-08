package notifier

import (
	"strings"
	"testing"
	"time"

	"github.com/exampleorg/cronwrap/internal/runner"
)

func successTemplateResult() runner.Result {
	return runner.Result{
		Command:   "echo hello",
		Success:   true,
		ExitCode:  0,
		Output:    "hello",
		Duration:  150 * time.Millisecond,
		StartedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func failureTemplateResult() runner.Result {
	return runner.Result{
		Command:   "false",
		Success:   false,
		ExitCode:  1,
		Output:    "",
		Err:       "exit status 1",
		Duration:  20 * time.Millisecond,
		StartedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func TestTemplateRenderer_DefaultTemplate_Success(t *testing.T) {
	r, err := NewTemplateRenderer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := r.Render(successTemplateResult())
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if !strings.Contains(out, "echo hello") {
		t.Errorf("expected command in output, got: %s", out)
	}
	if !strings.Contains(out, "success") {
		t.Errorf("expected status 'success', got: %s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected output snippet, got: %s", out)
	}
}

func TestTemplateRenderer_DefaultTemplate_Failure(t *testing.T) {
	r, _ := NewTemplateRenderer("")
	out, err := r.Render(failureTemplateResult())
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if !strings.Contains(out, "failure") {
		t.Errorf("expected status 'failure', got: %s", out)
	}
	if !strings.Contains(out, "exit status 1") {
		t.Errorf("expected error message, got: %s", out)
	}
}

func TestTemplateRenderer_CustomTemplate(t *testing.T) {
	r, err := NewTemplateRenderer("ALERT: {{.Command}} exited {{.ExitCode}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := r.Render(failureTemplateResult())
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	expected := "ALERT: false exited 1"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestTemplateRenderer_InvalidTemplate(t *testing.T) {
	_, err := NewTemplateRenderer("{{.Unclosed")
	if err == nil {
		t.Fatal("expected parse error for invalid template")
	}
}

func TestTemplateRenderer_DurationFormatted(t *testing.T) {
	r, _ := NewTemplateRenderer("{{.Duration}}")
	out, _ := r.Render(successTemplateResult())
	if !strings.Contains(out, "ms") {
		t.Errorf("expected millisecond duration, got: %s", out)
	}
}
