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
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
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
		t.Errorf("command = %q, want echo", cfg.Command)
	}
	if cfg.JobName != "test-job" {
		t.Errorf("job_name = %q, want test-job", cfg.JobName)
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
	path := writeTemp(t, "command: ls\nopsgenie:\n  api_key: original\n")
	t.Setenv("CRONWRAP_OPSGENIE_API_KEY", "env-key")
	t.Setenv("CRONWRAP_OPSGENIE_TEAM", "env-team")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OpsGenie.APIKey != "env-key" {
		t.Errorf("api_key = %q, want env-key", cfg.OpsGenie.APIKey)
	}
	if cfg.OpsGenie.Team != "env-team" {
		t.Errorf("team = %q, want env-team", cfg.OpsGenie.Team)
	}
}

func TestLoad_OpsGenieFields(t *testing.T) {
	path := writeTemp(t, "command: backup\nopsgenie:\n  api_key: abc123\n  team: sre\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OpsGenie.APIKey != "abc123" {
		t.Errorf("api_key = %q, want abc123", cfg.OpsGenie.APIKey)
	}
	if cfg.OpsGenie.Team != "sre" {
		t.Errorf("team = %q, want sre", cfg.OpsGenie.Team)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTemp(t, ": bad: yaml: [unclosed")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
