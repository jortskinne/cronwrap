package config

import (
	"testing"
	"time"
)

func TestBatchConfig_WindowDuration_Valid(t *testing.T) {
	b := BatchConfig{Window: "1m30s"}
	d, err := b.WindowDuration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 90*time.Second {
		t.Errorf("expected 90s, got %v", d)
	}
}

func TestBatchConfig_WindowDuration_Empty(t *testing.T) {
	b := BatchConfig{}
	d, err := b.WindowDuration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestBatchConfig_WindowDuration_Invalid(t *testing.T) {
	b := BatchConfig{Window: "notaduration"}
	_, err := b.WindowDuration()
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestApplyBatchDefaults_SetsWindow(t *testing.T) {
	b := BatchConfig{Enabled: true}
	ApplyBatchDefaults(&b)
	if b.Window != "30s" {
		t.Errorf("expected default window '30s', got %q", b.Window)
	}
}

func TestApplyBatchDefaults_PreservesWindow(t *testing.T) {
	b := BatchConfig{Enabled: true, Window: "2m"}
	ApplyBatchDefaults(&b)
	if b.Window != "2m" {
		t.Errorf("expected window '2m', got %q", b.Window)
	}
}

func TestApplyBatchDefaults_NegativeMaxSizeClamped(t *testing.T) {
	b := BatchConfig{Enabled: true, MaxSize: -5}
	ApplyBatchDefaults(&b)
	if b.MaxSize != 0 {
		t.Errorf("expected MaxSize 0 after clamp, got %d", b.MaxSize)
	}
}

func TestApplyBatchDefaults_DisabledNoOp(t *testing.T) {
	b := BatchConfig{Enabled: false, Window: "", MaxSize: -1}
	ApplyBatchDefaults(&b)
	// When disabled, defaults should not be applied.
	if b.Window != "" {
		t.Errorf("expected empty window when disabled, got %q", b.Window)
	}
	if b.MaxSize != -1 {
		t.Errorf("expected MaxSize unchanged when disabled, got %d", b.MaxSize)
	}
}
