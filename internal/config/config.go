package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all cronwrap configuration.
type Config struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	JobName string   `yaml:"job_name"`
	Timeout int      `yaml:"timeout_seconds"`

	Slack struct {
		WebhookURL string `yaml:"webhook_url"`
	} `yaml:"slack"`

	Email struct {
		SMTPHost string `yaml:"smtp_host"`
		SMTPPort int    `yaml:"smtp_port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
		To       string `yaml:"to"`
	} `yaml:"email"`

	Webhook struct {
		URL string `yaml:"url"`
	} `yaml:"webhook"`

	PagerDuty struct {
		RoutingKey string `yaml:"routing_key"`
	} `yaml:"pagerduty"`

	OpsGenie struct {
		APIKey string `yaml:"api_key"`
		Team   string `yaml:"team"`
	} `yaml:"opsgenie"`

	History struct {
		File       string `yaml:"file"`
		MaxRecords int    `yaml:"max_records"`
		MaxAgeDays int    `yaml:"max_age_days"`
	} `yaml:"history"`

	NotifyOnSuccess bool `yaml:"notify_on_success"`
}

// Load reads a YAML config file from path and applies defaults and env overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if cfg.Command == "" {
		return nil, fmt.Errorf("config: 'command' is required")
	}

	ApplyDefaults(&cfg)
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

// applyEnvOverrides replaces config values with environment variables when set.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("CRONWRAP_SLACK_WEBHOOK"); v != "" {
		cfg.Slack.WebhookURL = v
	}
	if v := os.Getenv("CRONWRAP_OPSGENIE_API_KEY"); v != "" {
		cfg.OpsGenie.APIKey = v
	}
	if v := os.Getenv("CRONWRAP_OPSGENIE_TEAM"); v != "" {
		cfg.OpsGenie.Team = v
	}
	if v := os.Getenv("CRONWRAP_PAGERDUTY_KEY"); v != "" {
		cfg.PagerDuty.RoutingKey = v
	}
	if v := os.Getenv("CRONWRAP_EMAIL_PASSWORD"); v != "" {
		cfg.Email.Password = v
	}
}
