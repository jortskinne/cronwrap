package config

import "time"

// RetryConfig controls how many times a notifier is retried on failure.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (1 means no retry).
	MaxAttempts int `yaml:"max_attempts"`
	// DelaySeconds is the pause between attempts.
	DelaySeconds int `yaml:"delay_seconds"`
}

// Delay returns the retry delay as a time.Duration.
func (r RetryConfig) Delay() time.Duration {
	return time.Duration(r.DelaySeconds) * time.Second
}

// ApplyRetryDefaults fills in zero-value fields with sensible defaults.
func ApplyRetryDefaults(r *RetryConfig) {
	if r.MaxAttempts <= 0 {
		r.MaxAttempts = 1
	}
	if r.DelaySeconds < 0 {
		r.DelaySeconds = 0
	}
}
