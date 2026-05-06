package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwrap-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, "command: echo\nargs: [hello]\njob_name: test-job\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "echo" {
		t.Errorf("expected command echo, got %s", cfg.Command)
	}
	if cfg.JobName != "test-job" {
		t.Errorf("expected job_name test-job, got %s", cfg.JobName)
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	path := writeTemp(t, "job_name: no-cmd\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	path := writeTemp(t, "command: echo\n")
	t.Setenv("CRONWRAP_SLACK_WEBHOOK", "https://hooks.slack.com/test")
	t.Setenv("CRONWRAP_PAGERDUTY_KEY", "pd-integration-key")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Slack.WebhookURL != "https://hooks.slack.com/test" {
		t.Errorf("slack webhook not overridden, got %s", cfg.Slack.WebhookURL)
	}
	if cfg.PagerDuty.IntegrationKey != "pd-integration-key" {
		t.Errorf("pagerduty key not overridden, got %s", cfg.PagerDuty.IntegrationKey)
	}
}

func TestLoad_PagerDutyConfig(t *testing.T) {
	path := writeTemp(t, "command: echo\npagerduty:\n  integration_key: my-key\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PagerDuty.IntegrationKey != "my-key" {
		t.Errorf("expected integration key my-key, got %s", cfg.PagerDuty.IntegrationKey)
	}
}

func TestLoad_NotifyOnSuccess(t *testing.T) {
	path := writeTemp(t, "command: echo\nnotify_on_success: true\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.NotifyOnSuccess {
		t.Error("expected notify_on_success to be true")
	}
}
