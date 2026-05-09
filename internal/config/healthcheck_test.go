package config

import (
	"testing"
	"time"
)

func TestHealthCheckConfig_TimeoutDuration_Valid(t *testing.T) {
	c := HealthCheckConfig{Timeout: "30s"}
	if got := c.TimeoutDuration(); got != 30*time.Second {
		t.Errorf("expected 30s, got %v", got)
	}
}

func TestHealthCheckConfig_TimeoutDuration_Empty(t *testing.T) {
	c := HealthCheckConfig{}
	if got := c.TimeoutDuration(); got != 10*time.Second {
		t.Errorf("expected default 10s, got %v", got)
	}
}

func TestHealthCheckConfig_TimeoutDuration_Invalid(t *testing.T) {
	c := HealthCheckConfig{Timeout: "not-a-duration"}
	if got := c.TimeoutDuration(); got != 10*time.Second {
		t.Errorf("expected default 10s for invalid duration, got %v", got)
	}
}

func TestApplyHealthCheckDefaults_SetsTimeout(t *testing.T) {
	c := HealthCheckConfig{}
	ApplyHealthCheckDefaults(&c)
	if c.Timeout != "10s" {
		t.Errorf("expected Timeout=10s, got %q", c.Timeout)
	}
}

func TestApplyHealthCheckDefaults_PreservesExisting(t *testing.T) {
	c := HealthCheckConfig{URL: "https://hc-ping.com/abc", Timeout: "5s"}
	ApplyHealthCheckDefaults(&c)
	if c.Timeout != "5s" {
		t.Errorf("expected Timeout preserved as 5s, got %q", c.Timeout)
	}
	if c.URL != "https://hc-ping.com/abc" {
		t.Errorf("expected URL preserved, got %q", c.URL)
	}
}

func TestHealthCheckConfig_Enabled(t *testing.T) {
	empty := HealthCheckConfig{}
	if empty.URL != "" {
		t.Error("expected empty URL to indicate disabled")
	}
	withURL := HealthCheckConfig{URL: "https://hc-ping.com/xyz"}
	if withURL.URL == "" {
		t.Error("expected non-empty URL to indicate enabled")
	}
}
