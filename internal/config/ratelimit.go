package config

import "time"

// RateLimitConfig controls how frequently notifications are sent
// for the same job within a rolling time window.
type RateLimitConfig struct {
	// Enabled turns rate limiting on or off.
	Enabled bool `toml:"enabled" yaml:"enabled"`

	// Interval is the minimum duration between notifications for the same job.
	// Expressed as a Go duration string, e.g. "5m", "1h".
	Interval string `toml:"interval" yaml:"interval"`
}

// IntervalDuration parses Interval into a time.Duration.
// Returns 0 if the field is empty or unparseable.
func (r RateLimitConfig) IntervalDuration() time.Duration {
	if r.Interval == "" {
		return 0
	}
	d, err := time.ParseDuration(r.Interval)
	if err != nil {
		return 0
	}
	return d
}

// ApplyRateLimitDefaults sets sensible defaults for an empty RateLimitConfig.
func ApplyRateLimitDefaults(cfg *RateLimitConfig) {
	if cfg.Interval == "" {
		cfg.Interval = "5m"
	}
}
