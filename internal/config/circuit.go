package config

import "time"

// CircuitBreakerConfig controls the circuit-breaker behaviour for notifiers.
type CircuitBreakerConfig struct {
	// Enabled toggles the circuit breaker (default: false).
	Enabled bool `yaml:"enabled"`

	// MaxFailures is the number of consecutive notifier errors before the
	// circuit opens. Must be >= 1.
	MaxFailures int `yaml:"max_failures"`

	// ResetTimeout is the duration the circuit stays open before allowing a
	// probe attempt. Accepts Go duration strings, e.g. "30s".
	ResetTimeout time.Duration `yaml:"reset_timeout"`
}

// ApplyCircuitBreakerDefaults fills zero-value fields with sensible defaults.
func ApplyCircuitBreakerDefaults(c *CircuitBreakerConfig) {
	if c.MaxFailures <= 0 {
		c.MaxFailures = 3
	}
	if c.ResetTimeout <= 0 {
		c.ResetTimeout = 30 * time.Second
	}
}
