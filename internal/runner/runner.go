package runner

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Result holds the outcome of a command execution.
type Result struct {
	Command   string
	Args      []string
	Stdout    string
	Stderr    string
	ExitCode  int
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
	Err       error
}

// Run executes the given command with optional arguments and a timeout.
// A timeout of 0 means no timeout.
func Run(ctx context.Context, command string, args []string, timeout time.Duration) *Result {
	result := &Result{
		Command:   command,
		Args:      args,
		StartedAt: time.Now(),
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	result.EndedAt = time.Now()
	result.Duration = result.EndedAt.Sub(result.StartedAt)
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	result.Err = err

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		result.ExitCode = -1
	}

	return result
}

// Success returns true if the command exited with code 0.
func (r *Result) Success() bool {
	return r.ExitCode == 0 && r.Err == nil
}
