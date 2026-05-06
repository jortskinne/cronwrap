package notifier

// NotifyOn controls which job outcomes trigger a notification.
type NotifyOn int

const (
	// NotifyOnFailure sends notifications only when a job fails (default).
	NotifyOnFailure NotifyOn = iota
	// NotifyOnAlways sends notifications for every job run.
	NotifyOnAlways
	// NotifyOnSuccess sends notifications only when a job succeeds.
	NotifyOnSuccess
)

// ParseNotifyOn converts a string value from config into a NotifyOn constant.
// Unrecognised values fall back to NotifyOnFailure.
func ParseNotifyOn(s string) NotifyOn {
	switch s {
	case "always":
		return NotifyOnAlways
	case "success":
		return NotifyOnSuccess
	default:
		return NotifyOnFailure
	}
}

// FilteredNotifier wraps a Notifier and only forwards calls that match the
// configured NotifyOn policy.
type FilteredNotifier struct {
	inner    Notifier
	notifyOn NotifyOn
}

// NewFilteredNotifier returns a Notifier that applies the given policy before
// delegating to inner.
func NewFilteredNotifier(inner Notifier, policy NotifyOn) *FilteredNotifier {
	return &FilteredNotifier{inner: inner, notifyOn: policy}
}

// Notify forwards the result to the inner Notifier only when the policy allows
// it, and returns nil otherwise.
func (f *FilteredNotifier) Notify(result JobResult) error {
	switch f.notifyOn {
	case NotifyOnAlways:
		return f.inner.Notify(result)
	case NotifyOnSuccess:
		if result.ExitCode == 0 {
			return f.inner.Notify(result)
		}
	default: // NotifyOnFailure
		if result.ExitCode != 0 {
			return f.inner.Notify(result)
		}
	}
	return nil
}
