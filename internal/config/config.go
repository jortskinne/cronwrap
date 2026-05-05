// Package config handles loading and validating cronwrap configuration
// from YAML files and environment variables.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all cronwrap configuration.
type Config struct {
	Command  string        `yaml:"command"`
	Timeout  time.Duration `yaml:"timeout"`
	Slack    SlackConfig   `yaml:"slack"`
	Email    EmailConfig   `yaml:"email"`
	History  HistoryConfig `yaml:"history"`
}

// SlackConfig holds Slack notifier settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

// EmailConfig holds SMTP notifier settings.
type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// HistoryConfig holds run history settings.
type HistoryConfig struct {
	Path    string `yaml:"path"`
	MaxRows int    `yaml:"max_rows"`
}

// Load reads a YAML config file from the given path and returns a Config.
// Environment variables prefixed with CRONWRAP_ override file values.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	applyEnvOverrides(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyEnvOverrides replaces zero-value fields with environment variable values.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("CRONWRAP_SLACK_WEBHOOK"); v != "" {
		cfg.Slack.WebhookURL = v
	}
	if v := os.Getenv("CRONWRAP_EMAIL_PASSWORD"); v != "" {
		cfg.Email.Password = v
	}
	if v := os.Getenv("CRONWRAP_HISTORY_PATH"); v != "" {
		cfg.History.Path = v
	}
}

// Validate checks that the configuration is consistent.
func (c *Config) Validate() error {
	if c.Command == "" {
		return fmt.Errorf("config: command must not be empty")
	}
	if c.Timeout < 0 {
		return fmt.Errorf("config: timeout must be non-negative")
	}
	if c.History.MaxRows < 0 {
		return fmt.Errorf("config: history.max_rows must be non-negative")
	}
	return nil
}
