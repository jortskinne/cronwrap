package config

import (
	"testing"
	"time"
)

func TestApplyCircuitBreakerDefaults_ZeroValues(t *testing.T) {
	c := &CircuitBreakerConfig{}
	ApplyCircuitBreakerDefaults(c)

	if c.MaxFailures != 3 {
		t.Errorf("expected MaxFailures=3, got %d", c.MaxFailures)
	}
	if c.ResetTimeout != 30*time.Second {
		t.Errorf("expected ResetTimeout=30s, got %v", c.ResetTimeout)
	}
}

func TestApplyCircuitBreakerDefaults_PreservesExisting(t *testing.T) {
	c := &CircuitBreakerConfig{
		MaxFailures:  5,
		ResetTimeout: 2 * time.Minute,
	}
	ApplyCircuitBreakerDefaults(c)

	if c.MaxFailures != 5 {
		t.Errorf("expected MaxFailures=5, got %d", c.MaxFailures)
	}
	if c.ResetTimeout != 2*time.Minute {
		t.Errorf("expected ResetTimeout=2m, got %v", c.ResetTimeout)
	}
}

func TestApplyCircuitBreakerDefaults_NegativeMaxFailures(t *testing.T) {
	c := &CircuitBreakerConfig{MaxFailures: -1}
	ApplyCircuitBreakerDefaults(c)

	if c.MaxFailures != 3 {
		t.Errorf("expected MaxFailures defaulted to 3, got %d", c.MaxFailures)
	}
}

func TestCircuitBreakerConfig_EnabledField(t *testing.T) {
	c := &CircuitBreakerConfig{Enabled: true}
	ApplyCircuitBreakerDefaults(c)

	if !c.Enabled {
		t.Error("expected Enabled=true to be preserved")
	}
}
