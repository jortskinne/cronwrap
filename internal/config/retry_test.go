package config

import (
	"testing"
	"time"
)

func TestRetryConfig_Delay(t *testing.T) {
	r := RetryConfig{DelaySeconds: 5}
	if got := r.Delay(); got != 5*time.Second {
		t.Fatalf("expected 5s, got %v", got)
	}
}

func TestRetryConfig_DelayZero(t *testing.T) {
	r := RetryConfig{DelaySeconds: 0}
	if got := r.Delay(); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestApplyRetryDefaults_ZeroValues(t *testing.T) {
	r := RetryConfig{}
	ApplyRetryDefaults(&r)
	if r.MaxAttempts != 1 {
		t.Errorf("expected MaxAttempts=1, got %d", r.MaxAttempts)
	}
	if r.DelaySeconds != 0 {
		t.Errorf("expected DelaySeconds=0, got %d", r.DelaySeconds)
	}
}

func TestApplyRetryDefaults_PreservesPositiveValues(t *testing.T) {
	r := RetryConfig{MaxAttempts: 4, DelaySeconds: 10}
	ApplyRetryDefaults(&r)
	if r.MaxAttempts != 4 {
		t.Errorf("expected MaxAttempts=4, got %d", r.MaxAttempts)
	}
	if r.DelaySeconds != 10 {
		t.Errorf("expected DelaySeconds=10, got %d", r.DelaySeconds)
	}
}

func TestApplyRetryDefaults_NegativeDelayClamped(t *testing.T) {
	r := RetryConfig{MaxAttempts: 2, DelaySeconds: -3}
	ApplyRetryDefaults(&r)
	if r.DelaySeconds != 0 {
		t.Errorf("expected DelaySeconds=0, got %d", r.DelaySeconds)
	}
}

func TestApplyRetryDefaults_NegativeAttemptsClamped(t *testing.T) {
	r := RetryConfig{MaxAttempts: -1}
	ApplyRetryDefaults(&r)
	if r.MaxAttempts != 1 {
		t.Errorf("expected MaxAttempts=1, got %d", r.MaxAttempts)
	}
}
