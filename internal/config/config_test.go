package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
command: echo hello
timeout: 30s
slack:
  webhook_url: https://hooks.slack.com/xxx
history:
  path: /tmp/history.jsonl
  max_rows: 100
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "echo hello" {
		t.Errorf("command = %q, want %q", cfg.Command, "echo hello")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.History.MaxRows != 100 {
		t.Errorf("max_rows = %d, want 100", cfg.History.MaxRows)
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	path := writeTemp(t, `timeout: 10s`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	path := writeTemp(t, `command: ls`)
	t.Setenv("CRONWRAP_SLACK_WEBHOOK", "https://env-webhook")
	t.Setenv("CRONWRAP_HISTORY_PATH", "/tmp/env-history.jsonl")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Slack.WebhookURL != "https://env-webhook" {
		t.Errorf("webhook = %q, want env override", cfg.Slack.WebhookURL)
	}
	if cfg.History.Path != "/tmp/env-history.jsonl" {
		t.Errorf("history path = %q, want env override", cfg.History.Path)
	}
}

func TestValidate_NegativeTimeout(t *testing.T) {
	cfg := &config.Config{Command: "ls", Timeout: -1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

func TestValidate_NegativeMaxRows(t *testing.T) {
	cfg := &config.Config{Command: "ls"}
	cfg.History.MaxRows = -5
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative max_rows")
	}
}
