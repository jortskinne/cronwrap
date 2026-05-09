package config

import "time"

// HealthCheckConfig holds configuration for ping-based health-check services
// such as healthchecks.io or Uptime Kuma.
type HealthCheckConfig struct {
	// URL is the base ping URL, e.g. "https://hc-ping.com/<uuid>".
	// The notifier appends "/success" or "/fail" automatically.
	URL string `yaml:"url"`

	// Timeout is the HTTP request timeout as a duration string (e.g. "10s").
	// Defaults to "10s" when empty.
	Timeout string `yaml:"timeout"`
}

// TimeoutDuration parses Timeout into a time.Duration.
// Returns the default of 10 s when the field is empty or unparseable.
func (h HealthCheckConfig) TimeoutDuration() time.Duration {
	if h.Timeout == "" {
		return 10 * time.Second
	}
	d, err := time.ParseDuration(h.Timeout)
	if err != nil {
		return 10 * time.Second
	}
	return d
}

// ApplyHealthCheckDefaults fills in zero-value fields with sensible defaults.
func ApplyHealthCheckDefaults(c *HealthCheckConfig) {
	if c.Timeout == "" {
		c.Timeout = "10s"
	}
}
