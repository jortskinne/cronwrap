package config_test

import (
	"testing"

	"github.com/yourorg/cronwrap/internal/config"
)

func TestSentryConfig_Enabled_WithDSN(t *testing.T) {
	c := config.SentryConfig{DSN: "https://key@sentry.io/123"}
	if !c.Enabled() {
		t.Error("expected Enabled()=true when DSN is set")
	}
}

func TestSentryConfig_Enabled_NoDSN(t *testing.T) {
	c := config.SentryConfig{}
	if c.Enabled() {
		t.Error("expected Enabled()=false when DSN is empty")
	}
}

func TestApplySentryDefaults_SetsEnvironment(t *testing.T) {
	c := &config.SentryConfig{DSN: "https://key@sentry.io/1"}
	config.ApplySentryDefaults(c)
	if c.Environment != "production" {
		t.Errorf("expected environment=production, got %q", c.Environment)
	}
}

func TestApplySentryDefaults_PreservesEnvironment(t *testing.T) {
	c := &config.SentryConfig{
		DSN:         "https://key@sentry.io/1",
		Environment: "staging",
	}
	config.ApplySentryDefaults(c)
	if c.Environment != "staging" {
		t.Errorf("expected environment=staging, got %q", c.Environment)
	}
}

func TestApplySentryDefaults_EmptyRelease(t *testing.T) {
	c := &config.SentryConfig{DSN: "https://key@sentry.io/1"}
	config.ApplySentryDefaults(c)
	// Release has no default — should remain empty
	if c.Release != "" {
		t.Errorf("expected release to remain empty, got %q", c.Release)
	}
}
