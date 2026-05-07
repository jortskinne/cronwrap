package config

import "time"

// DedupeConfig controls the deduplication notifier behaviour.
type DedupeConfig struct {
	// Enabled turns deduplication on or off.
	Enabled bool `toml:"enabled" yaml:"enabled"`

	// Window is the duration during which identical notifications are
	// suppressed. Accepts Go duration strings such as "10m" or "1h".
	Window time.Duration `toml:"window" yaml:"window"`
}

// ApplyDedupeDefaults fills in zero values with sensible defaults.
func ApplyDedupeDefaults(c *DedupeConfig) {
	if c.Window <= 0 {
		c.Window = 5 * time.Minute
	}
}
