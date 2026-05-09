package config

// SentryConfig holds configuration for the Sentry notifier.
type SentryConfig struct {
	// DSN is the Sentry ingest endpoint URL (required to enable).
	DSN string `toml:"dsn" yaml:"dsn"`

	// Environment labels the Sentry event (e.g. "production", "staging").
	Environment string `toml:"environment" yaml:"environment"`

	// Release is an optional release/version string attached to events.
	Release string `toml:"release" yaml:"release"`
}

// Enabled returns true when a DSN has been configured.
func (s SentryConfig) Enabled() bool {
	return s.DSN != ""
}

// ApplySentryDefaults sets sensible defaults for optional Sentry fields.
func ApplySentryDefaults(c *SentryConfig) {
	if c.Environment == "" {
		c.Environment = "production"
	}
}
