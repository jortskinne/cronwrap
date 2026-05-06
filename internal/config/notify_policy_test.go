package config

import (
	"testing"

	"github.com/cronwrap/internal/notifier"
)

func TestNotifyPolicy_ToNotifyOn_Failure(t *testing.T) {
	p := NotifyPolicy{On: "failure"}
	if got := p.ToNotifyOn(); got != notifier.NotifyOnFailure {
		t.Errorf("expected NotifyOnFailure, got %v", got)
	}
}

func TestNotifyPolicy_ToNotifyOn_Always(t *testing.T) {
	p := NotifyPolicy{On: "always"}
	if got := p.ToNotifyOn(); got != notifier.NotifyOnAlways {
		t.Errorf("expected NotifyOnAlways, got %v", got)
	}
}

func TestNotifyPolicy_ToNotifyOn_Success(t *testing.T) {
	p := NotifyPolicy{On: "success"}
	if got := p.ToNotifyOn(); got != notifier.NotifyOnSuccess {
		t.Errorf("expected NotifyOnSuccess, got %v", got)
	}
}

func TestNotifyPolicy_ToNotifyOn_Empty(t *testing.T) {
	p := NotifyPolicy{}
	// Empty string should fall back to failure semantics.
	if got := p.ToNotifyOn(); got != notifier.NotifyOnFailure {
		t.Errorf("expected NotifyOnFailure for empty On, got %v", got)
	}
}

func TestApplyNotifyPolicyDefaults_SetsFailure(t *testing.T) {
	p := NotifyPolicy{}
	ApplyNotifyPolicyDefaults(&p)
	if p.On != "failure" {
		t.Errorf("expected 'failure', got %q", p.On)
	}
}

func TestApplyNotifyPolicyDefaults_PreservesExisting(t *testing.T) {
	p := NotifyPolicy{On: "always"}
	ApplyNotifyPolicyDefaults(&p)
	if p.On != "always" {
		t.Errorf("expected 'always' to be preserved, got %q", p.On)
	}
}
