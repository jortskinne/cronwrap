package config

import "time"

// BatchConfig controls how notifications are batched before delivery.
type BatchConfig struct {
	// Enabled turns batching on or off.
	Enabled bool `toml:"enabled" yaml:"enabled"`

	// Window is the maximum time to accumulate results before flushing,
	// expressed as a Go duration string (e.g. "30s", "2m").
	Window string `toml:"window" yaml:"window"`

	// MaxSize triggers an immediate flush when the buffer reaches this count.
	// Zero disables size-based flushing.
	MaxSize int `toml:"max_size" yaml:"max_size"`
}

// WindowDuration parses Window into a time.Duration.
// Returns 0 and no error when Window is empty.
func (b BatchConfig) WindowDuration() (time.Duration, error) {
	if b.Window == "" {
		return 0, nil
	}
	return time.ParseDuration(b.Window)
}

// ApplyBatchDefaults fills in sensible defaults for a BatchConfig.
func ApplyBatchDefaults(b *BatchConfig) {
	if !b.Enabled {
		return
	}
	if b.Window == "" {
		b.Window = "30s"
	}
	if b.MaxSize < 0 {
		b.MaxSize = 0
	}
}
