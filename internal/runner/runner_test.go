package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/runner"
)

func TestRun_SimpleCommand(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "echo-job", "echo hello", 0)

	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if strings.TrimSpace(res.Stdout) != "hello" {
		t.Errorf("expected stdout 'hello', got %q", res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
	if res.JobName != "echo-job" {
		t.Errorf("expected job name 'echo-job', got %q", res.JobName)
	}
}

func TestRun_NonZeroExit(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "fail-job", "exit 2", 0)

	if res.ExitCode != 2 {
		t.Errorf("expected exit code 2, got %d", res.ExitCode)
	}
	if res.Error == nil {
		t.Error("expected non-nil error for non-zero exit")
	}
}

func TestRun_Timeout(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "slow-job", "sleep 10", 100*time.Millisecond)

	if res.ExitCode != -1 {
		t.Errorf("expected exit code -1 on timeout, got %d", res.ExitCode)
	}
	if res.Error == nil || !strings.Contains(res.Error.Error(), "timed out") {
		t.Errorf("expected timeout error, got %v", res.Error)
	}
}

func TestRun_StderrCaptured(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "stderr-job", "echo errout >&2", 0)

	if strings.TrimSpace(res.Stderr) != "errout" {
		t.Errorf("expected stderr 'errout', got %q", res.Stderr)
	}
}

func TestRun_DurationRecorded(t *testing.T) {
	r := runner.New(5 * time.Second)
	res := r.Run(context.Background(), "dur-job", "sleep 0.05", 0)

	if res.Duration < 50*time.Millisecond {
		t.Errorf("expected duration >= 50ms, got %s", res.Duration)
	}
}
