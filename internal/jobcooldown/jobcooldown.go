// Package jobcooldown enforces a minimum interval between successive runs
// of the same job, preventing runaway re-execution after failures or drift.
package jobcooldown

import (
	"fmt"
	"sync"
	"time"
)

// Cooldown tracks the last run time per job and enforces a minimum gap.
type Cooldown struct {
	mu       sync.Mutex
	lastRun  map[string]time.Time
	minGap   time.Duration
	nowFn    func() time.Time
}

// New creates a Cooldown that enforces minGap between successive runs.
func New(minGap time.Duration) *Cooldown {
	return &Cooldown{
		lastRun: make(map[string]time.Time),
		minGap:  minGap,
		nowFn:   time.Now,
	}
}

// Allow returns nil if the job may run now, or an error describing how long
// the caller must wait before the cooldown expires.
func (c *Cooldown) Allow(jobName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.nowFn()
	last, seen := c.lastRun[jobName]
	if !seen {
		return nil
	}

	elapsed := now.Sub(last)
	if elapsed >= c.minGap {
		return nil
	}

	remaining := c.minGap - elapsed
	return fmt.Errorf("job %q is cooling down: %s remaining", jobName, remaining.Round(time.Millisecond))
}

// Record marks jobName as having just run at the current time.
func (c *Cooldown) Record(jobName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastRun[jobName] = c.nowFn()
}

// Reset clears the cooldown state for jobName, allowing it to run immediately.
func (c *Cooldown) Reset(jobName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.lastRun, jobName)
}

// LastRun returns the last recorded run time for jobName and whether it exists.
func (c *Cooldown) LastRun(jobName string) (time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.lastRun[jobName]
	return t, ok
}
