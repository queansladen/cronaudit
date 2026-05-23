// Package jobmeta tracks per-job metadata such as average duration,
// last exit code, and run count for use in reporting and alerting.
package jobmeta

import (
	"sync"
	"time"
)

// Meta holds aggregated metadata for a single job.
type Meta struct {
	JobName      string
	RunCount     int
	FailCount    int
	DriftCount   int
	TotalDuration time.Duration
	LastDuration  time.Duration
	LastExitCode  int
	LastRunAt     time.Time
}

// AvgDuration returns the mean run duration across all recorded runs.
func (m *Meta) AvgDuration() time.Duration {
	if m.RunCount == 0 {
		return 0
	}
	return m.TotalDuration / time.Duration(m.RunCount)
}

// Tracker stores and updates job metadata in memory.
type Tracker struct {
	mu   sync.RWMutex
	jobs map[string]*Meta
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{jobs: make(map[string]*Meta)}
}

// Record updates metadata for the named job with the result of a single run.
func (t *Tracker) Record(name string, exitCode int, duration time.Duration, drifted bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	m, ok := t.jobs[name]
	if !ok {
		m = &Meta{JobName: name}
		t.jobs[name] = m
	}

	m.RunCount++
	m.LastExitCode = exitCode
	m.LastDuration = duration
	m.TotalDuration += duration
	m.LastRunAt = time.Now()

	if exitCode != 0 {
		m.FailCount++
	}
	if drifted {
		m.DriftCount++
	}
}

// Get returns a copy of the metadata for the named job and whether it exists.
func (t *Tracker) Get(name string) (Meta, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m, ok := t.jobs[name]
	if !ok {
		return Meta{}, false
	}
	return *m, true
}

// All returns a snapshot of metadata for every tracked job.
func (t *Tracker) All() []Meta {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Meta, 0, len(t.jobs))
	for _, m := range t.jobs {
		out = append(out, *m)
	}
	return out
}
