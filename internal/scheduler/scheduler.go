// Package scheduler wires together the runner, store, differ, and notifier
// to execute cron jobs on their configured intervals.
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yourorg/cronaudit/internal/config"
	"github.com/yourorg/cronaudit/internal/differ"
	"github.com/yourorg/cronaudit/internal/notifier"
	"github.com/yourorg/cronaudit/internal/runner"
	"github.com/yourorg/cronaudit/internal/store"
)

// Scheduler runs each configured job on its interval.
type Scheduler struct {
	cfg      *config.Config
	runner   *runner.Runner
	store    *store.Store
	notifier *notifier.Notifier
}

// New creates a Scheduler from the provided config and dependencies.
func New(cfg *config.Config, r *runner.Runner, s *store.Store, n *notifier.Notifier) *Scheduler {
	return &Scheduler{
		cfg:      cfg,
		runner:   r,
		store:    s,
		notifier: n,
	}
}

// Run starts all job loops and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, job := range s.cfg.Jobs {
		wg.Add(1)
		go func(j config.Job) {
			defer wg.Done()
			s.runLoop(ctx, j)
		}(job)
	}
	wg.Wait()
}

// runLoop executes a single job on its interval until ctx is done.
func (s *Scheduler) runLoop(ctx context.Context, job config.Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run immediately on start, then on each tick.
	s.execute(ctx, job)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.execute(ctx, job)
		}
	}
}

// execute runs the job once, stores the result, diffs against the previous run,
// and notifies accordingly.
func (s *Scheduler) execute(ctx context.Context, job config.Job) {
	prev, _ := s.store.Latest(job.Name)

	result, err := s.runner.Run(ctx, job)
	if err != nil {
		log.Printf("scheduler: job %q run error: %v", job.Name, err)
		return
	}

	if err := s.store.Save(job.Name, result); err != nil {
		log.Printf("scheduler: job %q store error: %v", job.Name, err)
	}

	diff := differ.Compare(prev, result)
	if diff.HasChanges {
		s.notifier.NotifyDrift(job.Name, diff)
	} else {
		s.notifier.Notify(notifier.Event{
			JobName:   job.Name,
			Kind:      notifier.KindInfo,
			Timestamp: result.StartedAt,
			Message:   "no drift detected",
		})
	}
}
