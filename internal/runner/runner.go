package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Result holds the output and metadata from a single job execution.
type Result struct {
	JobName   string
	Command   string
	Stdout    string
	Stderr    string
	ExitCode  int
	StartedAt time.Time
	Duration  time.Duration
	Error     error
}

// Runner executes shell commands with a configurable timeout.
type Runner struct {
	DefaultTimeout time.Duration
}

// New returns a Runner with the given default timeout.
func New(defaultTimeout time.Duration) *Runner {
	return &Runner{DefaultTimeout: defaultTimeout}
}

// Run executes the given command string under /bin/sh and returns a Result.
// If timeout is zero, the runner's DefaultTimeout is used.
func (r *Runner) Run(ctx context.Context, jobName, command string, timeout time.Duration) Result {
	if timeout == 0 {
		timeout = r.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	result := Result{
		JobName:   jobName,
		Command:   command,
		StartedAt: start,
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("command timed out after %s", timeout)
			result.ExitCode = -1
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
