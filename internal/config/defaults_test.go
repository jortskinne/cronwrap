package config_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/config"
)

func TestApplyDefaults_AllZero(t *testing.T) {
	cfg := &config.Config{Command: "echo"}
	config.ApplyDefaults(cfg)

	if cfg.Timeout != config.DefaultTimeout {
		t.Errorf("timeout = %v, want %v", cfg.Timeout, config.DefaultTimeout)
	}
	if cfg.History.Path != config.DefaultHistoryPath {
		t.Errorf("history.path = %q, want %q", cfg.History.Path, config.DefaultHistoryPath)
	}
	if cfg.History.MaxRows != config.DefaultMaxRows {
		t.Errorf("history.max_rows = %d, want %d", cfg.History.MaxRows, config.DefaultMaxRows)
	}
	if cfg.Email.SMTPPort != config.DefaultSMTPPort {
		t.Errorf("email.smtp_port = %d, want %d", cfg.Email.SMTPPort, config.DefaultSMTPPort)
	}
}

func TestApplyDefaults_PreservesExisting(t *testing.T) {
	cfg := &config.Config{
		Command: "ls",
		Timeout: 5 * time.Second,
	}
	cfg.History.Path = "/custom/path.jsonl"
	cfg.History.MaxRows = 42
	cfg.Email.SMTPPort = 465

	config.ApplyDefaults(cfg)

	if cfg.Timeout != 5*time.Second {
		t.Errorf("timeout overwritten: got %v", cfg.Timeout)
	}
	if cfg.History.Path != "/custom/path.jsonl" {
		t.Errorf("history.path overwritten: got %q", cfg.History.Path)
	}
	if cfg.History.MaxRows != 42 {
		t.Errorf("history.max_rows overwritten: got %d", cfg.History.MaxRows)
	}
	if cfg.Email.SMTPPort != 465 {
		t.Errorf("smtp_port overwritten: got %d", cfg.Email.SMTPPort)
	}
}
