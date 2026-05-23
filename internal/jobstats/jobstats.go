// Package jobstats aggregates per-job execution statistics for reporting
// and threshold-based alerting.
package jobstats

import (
	"sync"
	"time"
)

// Stats holds cumulative statistics for a single job.
type Stats struct {
	JobName       string
	TotalRuns     int
	Failures      int
	Drifts        int
	LastRun       time.Time
	TotalDuration time.Duration
}

// AvgDuration returns the mean execution duration across all runs.
// Returns 0 if no runs have been recorded.
func (s *Stats) AvgDuration() time.Duration {
	if s.TotalRuns == 0 {
		return 0
	}
	return s.TotalDuration / time.Duration(s.TotalRuns)
}

// SuccessRate returns the fraction of runs that succeeded (0.0–1.0).
func (s *Stats) SuccessRate() float64 {
	if s.TotalRuns == 0 {
		return 0
	}
	return float64(s.TotalRuns-s.Failures) / float64(s.TotalRuns)
}

// Tracker maintains a thread-safe map of job statistics.
type Tracker struct {
	mu    sync.RWMutex
	stats map[string]*Stats
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{stats: make(map[string]*Stats)}
}

// Record incorporates the result of a single job execution into the tracker.
func (t *Tracker) Record(jobName string, duration time.Duration, failed, drifted bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, ok := t.stats[jobName]
	if !ok {
		s = &Stats{JobName: jobName}
		t.stats[jobName] = s
	}

	s.TotalRuns++
	s.TotalDuration += duration
	s.LastRun = time.Now()
	if failed {
		s.Failures++
	}
	if drifted {
		s.Drifts++
	}
}

// Get returns a copy of the Stats for the named job and whether it was found.
func (t *Tracker) Get(jobName string) (Stats, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.stats[jobName]
	if !ok {
		return Stats{}, false
	}
	return *s, true
}

// All returns a snapshot of statistics for every tracked job.
func (t *Tracker) All() []Stats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Stats, 0, len(t.stats))
	for _, s := range t.stats {
		out = append(out, *s)
	}
	return out
}
