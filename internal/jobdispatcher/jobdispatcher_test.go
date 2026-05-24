package jobdispatcher_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/jobdispatcher"
	"github.com/example/cronaudit/internal/jobqueue"
	"github.com/example/cronaudit/internal/joblock"
)

func newDispatcher(t *testing.T, workers int) (*jobdispatcher.Dispatcher, *jobqueue.Queue) {
	t.Helper()
	q, err := jobqueue.New(64)
	if err != nil {
		t.Fatalf("jobqueue.New: %v", err)
	}
	l := joblock.New()
	d, err := jobdispatcher.New(q, l, workers)
	if err != nil {
		t.Fatalf("jobdispatcher.New: %v", err)
	}
	return d, q
}

func TestNew_InvalidWorkers(t *testing.T) {
	q, _ := jobqueue.New(8)
	_, err := jobdispatcher.New(q, joblock.New(), 0)
	if err == nil {
		t.Fatal("expected error for workers=0")
	}
}

func TestDispatch_JobIsExecuted(t *testing.T) {
	d, q := newDispatcher(t, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d.Start(ctx)
	defer d.Stop()

	var ran atomic.Bool
	job := jobdispatcher.Job{
		Name:     "test-job",
		Priority: 1,
		Run: func(_ context.Context) error {
			ran.Store(true)
			return nil
		},
	}
	if err := q.Enqueue(job, job.Priority); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && !ran.Load() {
		time.Sleep(10 * time.Millisecond)
	}
	if !ran.Load() {
		t.Fatal("job was not executed within deadline")
	}
}

func TestDispatch_ErrorSurfaced(t *testing.T) {
	d, q := newDispatcher(t, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d.Start(ctx)

	sentinel := errors.New("boom")
	job := jobdispatcher.Job{
		Name:     "failing-job",
		Priority: 1,
		Run: func(_ context.Context) error { return sentinel },
	}
	_ = q.Enqueue(job, job.Priority)

	var got error
	select {
	case got = <-d.Errors():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for error")
	}
	cancel()
	d.Stop()
	if !errors.Is(got, sentinel) {
		t.Fatalf("expected sentinel error, got %v", got)
	}
}

func TestDispatch_ConcurrentJobs(t *testing.T) {
	d, q := newDispatcher(t, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	d.Start(ctx)
	defer d.Stop()

	var count atomic.Int32
	for i := 0; i < 8; i++ {
		name := fmt.Sprintf("job-%d", i)
		job := jobdispatcher.Job{
			Name:     name,
			Priority: 1,
			Run: func(_ context.Context) error { count.Add(1); return nil },
		}
		_ = q.Enqueue(job, job.Priority)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && count.Load() < 8 {
		time.Sleep(20 * time.Millisecond)
	}
	if count.Load() < 8 {
		t.Fatalf("expected 8 jobs executed, got %d", count.Load())
	}
}
