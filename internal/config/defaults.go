package config

import "time"

const (
	// DefaultTimeout is used when no timeout is specified in the config file.
	DefaultTimeout = 30 * time.Minute

	// DefaultHistoryPath is the default location for the run-history file.
	DefaultHistoryPath = "~/.cronwrap/history.jsonl"

	// DefaultMaxRows is the default maximum number of history records kept.
	DefaultMaxRows = 500

	// DefaultSMTPPort is the default SMTP submission port.
	DefaultSMTPPort = 587
)

// ApplyDefaults fills in zero values with sensible defaults.
// It is called automatically by Load but can also be used in tests.
func ApplyDefaults(cfg *Config) {
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.History.Path == "" {
		cfg.History.Path = DefaultHistoryPath
	}
	if cfg.History.MaxRows == 0 {
		cfg.History.MaxRows = DefaultMaxRows
	}
	if cfg.Email.SMTPPort == 0 {
		cfg.Email.SMTPPort = DefaultSMTPPort
	}
}
