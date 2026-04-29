// Package notifier provides interfaces and types for sending job result
// notifications via various channels such as Slack and email.
package notifier

import "time"

// JobResult contains the outcome of a cron job execution.
type JobResult struct {
	// JobName is the human-readable name or command of the job.
	JobName string
	// Success indicates whether the job exited with code 0.
	Success bool
	// ExitCode is the process exit code.
	ExitCode int
	// Output is the combined stdout captured during execution.
	Output string
	// Error is the stderr output captured during execution.
	Error string
	// Duration is how long the job ran.
	Duration time.Duration
	// StartedAt is when the job began.
	StartedAt time.Time
}

// Notifier is implemented by any type that can send a job result notification.
type Notifier interface {
	Notify(result JobResult) error
}

// Multi fans out a notification to multiple Notifier implementations.
// All notifiers are called; errors are collected and the first non-nil
// error is returned.
type Multi struct {
	notifiers []Notifier
}

// NewMulti creates a Multi notifier from the provided list.
func NewMulti(nn ...Notifier) *Multi {
	return &Multi{notifiers: nn}
}

// Notify sends the result to every registered notifier.
// It returns the first error encountered, but always attempts all notifiers.
func (m *Multi) Notify(result JobResult) error {
	var firstErr error
	for _, n := range m.notifiers {
		if err := n.Notify(result); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
