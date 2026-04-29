// Package runner provides functionality to execute shell commands as cron job
// tasks, capturing their stdout and stderr output, exit codes, and timing
// information.
//
// Basic usage:
//
//	result := runner.Run(ctx, "backup.sh", nil, 30*time.Minute)
//	if !result.Success() {
//	    // handle failure: notify via Slack/email, store in history, etc.
//	    fmt.Printf("job failed (exit %d): %s\n", result.ExitCode, result.Stderr)
//	}
//
// The Result struct contains all information needed for downstream reporting
// and history tracking:
//   - Stdout / Stderr: full captured output
//   - ExitCode: process exit status (-1 if the process could not start)
//   - StartedAt / EndedAt / Duration: timing metadata
//   - Err: underlying Go error (e.g. context.DeadlineExceeded on timeout)
package runner
