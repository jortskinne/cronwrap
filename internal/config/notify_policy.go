package config

import "github.com/cronwrap/internal/notifier"

// NotifyPolicy holds the parsed notification-filter configuration for a job.
type NotifyPolicy struct {
	// On determines which outcomes trigger a notification.
	// Accepted values: "failure" (default), "success", "always".
	On string `toml:"on" yaml:"on"`
}

// ToNotifyOn converts the string policy value into the typed notifier constant.
func (p NotifyPolicy) ToNotifyOn() notifier.NotifyOn {
	return notifier.ParseNotifyOn(p.On)
}

// ApplyNotifyPolicyDefaults fills in zero-value NotifyPolicy fields with
// sensible defaults so callers never have to guard against empty strings.
func ApplyNotifyPolicyDefaults(p *NotifyPolicy) {
	if p.On == "" {
		p.On = "failure"
	}
}
