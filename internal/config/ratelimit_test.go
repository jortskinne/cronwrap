package config

import (
	"testing"
	"time"
)

func TestRateLimitConfig_IntervalDuration_Valid(t *testing.T) {
	cfg := RateLimitConfig{Interval: "10m"}
	got := cfg.IntervalDuration()
	if got != 10*time.Minute {
		t.Errorf("expected 10m, got %v", got)
	}
}

func TestRateLimitConfig_IntervalDuration_Empty(t *testing.T) {
	cfg := RateLimitConfig{}
	if d := cfg.IntervalDuration(); d != 0 {
		t.Errorf("expected 0 for empty interval, got %v", d)
	}
}

func TestRateLimitConfig_IntervalDuration_Invalid(t *testing.T) {
	cfg := RateLimitConfig{Interval: "not-a-duration"}
	if d := cfg.IntervalDuration(); d != 0 {
		t.Errorf("expected 0 for invalid duration, got %v", d)
	}
}

func TestApplyRateLimitDefaults_SetsInterval(t *testing.T) {
	cfg := &RateLimitConfig{}
	ApplyRateLimitDefaults(cfg)
	if cfg.Interval != "5m" {
		t.Errorf("expected default interval '5m', got %q", cfg.Interval)
	}
}

func TestApplyRateLimitDefaults_PreservesExisting(t *testing.T) {
	cfg := &RateLimitConfig{Interval: "1h"}
	ApplyRateLimitDefaults(cfg)
	if cfg.Interval != "1h" {
		t.Errorf("expected preserved interval '1h', got %q", cfg.Interval)
	}
}

func TestRateLimitConfig_EnabledField(t *testing.T) {
	cfg := RateLimitConfig{Enabled: true, Interval: "30s"}
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.IntervalDuration() != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.IntervalDuration())
	}
}
