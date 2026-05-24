// Package jobthrottle limits the maximum number of concurrently running
// instances of any single cron job, preventing resource exhaustion when
// jobs pile up due to slow execution or scheduling overlap.
package jobthrottle

import (
	"errors"
	"fmt"
	"sync"
)

// ErrThrottled is returned when a job has reached its concurrency limit.
var ErrThrottled = errors.New("job throttled: concurrency limit reached")

// Throttle tracks in-flight job counts and enforces per-job concurrency limits.
type Throttle struct {
	mu      sync.Mutex
	max     int
	counts  map[string]int
}

// New creates a Throttle that allows at most maxConcurrent simultaneous
// executions per job name. Panics if maxConcurrent < 1.
func New(maxConcurrent int) *Throttle {
	if maxConcurrent < 1 {
		panic("jobthrottle: maxConcurrent must be >= 1")
	}
	return &Throttle{
		max:    maxConcurrent,
		counts: make(map[string]int),
	}
}

// Acquire attempts to reserve a slot for the given job.
// Returns ErrThrottled if the concurrency limit is already reached.
func (t *Throttle) Acquire(job string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.counts[job] >= t.max {
		return fmt.Errorf("%w: job=%s current=%d max=%d",
			ErrThrottled, job, t.counts[job], t.max)
	}
	t.counts[job]++
	return nil
}

// Release decrements the in-flight count for the given job.
// It is safe to call Release even if the count is already zero.
func (t *Throttle) Release(job string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.counts[job] > 0 {
		t.counts[job]--
	}
}

// InFlight returns the current number of in-flight executions for a job.
func (t *Throttle) InFlight(job string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.counts[job]
}
