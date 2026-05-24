// Package jobdispatcher routes incoming job execution requests to available
// workers, respecting concurrency limits and priority ordering.
package jobdispatcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/example/cronaudit/internal/jobqueue"
	"github.com/example/cronaudit/internal/joblock"
)

// Job represents a dispatchable unit of work.
type Job struct {
	Name     string
	Priority int
	Run      func(ctx context.Context) error
}

// Dispatcher fans out queued jobs to a pool of worker goroutines.
type Dispatcher struct {
	mu       sync.Mutex
	queue    *jobqueue.Queue
	locker   *joblock.Locker
	workers  int
	cancelFn context.CancelFunc
	wg       sync.WaitGroup
	errors   chan error
}

// New creates a Dispatcher backed by the given queue and locker.
// workers controls the maximum number of concurrent job executions.
func New(q *jobqueue.Queue, l *joblock.Locker, workers int) (*Dispatcher, error) {
	if workers < 1 {
		return nil, fmt.Errorf("jobdispatcher: workers must be >= 1, got %d", workers)
	}
	return &Dispatcher{
		queue:   q,
		locker:  l,
		workers: workers,
		errors:  make(chan error, 64),
	}, nil
}

// Start launches worker goroutines that drain the queue until ctx is cancelled.
func (d *Dispatcher) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	d.cancelFn = cancel
	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go d.loop(workerCtx)
	}
}

// Stop signals all workers to finish and waits for them to drain.
func (d *Dispatcher) Stop() {
	if d.cancelFn != nil {
		d.cancelFn()
	}
	d.wg.Wait()
	close(d.errors)
}

// Errors returns a channel of non-fatal errors emitted during dispatch.
func (d *Dispatcher) Errors() <-chan error { return d.errors }

func (d *Dispatcher) loop(ctx context.Context) {
	defer d.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		item, err := d.queue.Dequeue()
		if err != nil {
			// queue empty — yield and retry
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		j, ok := item.(Job)
		if !ok {
			continue
		}
		if !d.locker.Acquire(j.Name) {
			// already running — re-enqueue at same priority
			_ = d.queue.Enqueue(item, j.Priority)
			continue
		}
		if err := j.Run(ctx); err != nil {
			select {
			case d.errors <- fmt.Errorf("job %q: %w", j.Name, err):
			default:
			}
		}
		d.locker.Release(j.Name)
	}
}
