package scheduler_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/config"
	"github.com/yourorg/cronaudit/internal/notifier"
	"github.com/yourorg/cronaudit/internal/runner"
	"github.com/yourorg/cronaudit/internal/scheduler"
	"github.com/yourorg/cronaudit/internal/store"
)

func newTestScheduler(t *testing.T, jobs []config.Job) (*scheduler.Scheduler, *store.Store) {
	t.Helper()
	dir := t.TempDir()

	cfg := &config.Config{Jobs: jobs}
	r := runner.New()
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	n := notifier.New(os.Stdout)
	return scheduler.New(cfg, r, s, n), s
}

func TestScheduler_RunsJobAndStoresResult(t *testing.T) {
	jobs := []config.Job{
		{Name: "echo-job", Command: "echo hello", Interval: 50 * time.Millisecond, Timeout: 5 * time.Second},
	}
	sched, st := newTestScheduler(t, jobs)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	sched.Run(ctx)

	result, err := st.Latest("echo-job")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if result == nil {
		t.Fatal("expected a stored result, got nil")
	}
	if result.Output != "hello\n" {
		t.Errorf("unexpected output: %q", result.Output)
	}
}

func TestScheduler_DetectsDrift(t *testing.T) {
	// Use a counter file so output changes between runs.
	counterFile := t.TempDir() + "/count.txt"
	if err := os.WriteFile(counterFile, []byte("0"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := "sh -c 'cat " + counterFile + "; echo 1 > " + counterFile + "'" 
	jobs := []config.Job{
		{Name: "drift-job", Command: cmd, Interval: 60 * time.Millisecond, Timeout: 5 * time.Second},
	}
	sched, st := newTestScheduler(t, jobs)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	sched.Run(ctx)

	result, err := st.Latest("drift-job")
	if err != nil || result == nil {
		t.Fatalf("expected stored result: %v", err)
	}
}

func TestScheduler_CancelStopsAllJobs(t *testing.T) {
	jobs := []config.Job{
		{Name: "job-a", Command: "echo a", Interval: 10 * time.Millisecond, Timeout: 5 * time.Second},
		{Name: "job-b", Command: "echo b", Interval: 10 * time.Millisecond, Timeout: 5 * time.Second},
	}
	sched, _ := newTestScheduler(t, jobs)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sched.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}
