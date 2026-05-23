// Package jobhistory provides a rolling window summary of recent job
// executions, tracking success/failure counts and last-seen timestamps.
package jobhistory

import (
	"sync"
	"time"
)

// Entry holds aggregated history for a single job.
type Entry struct {
	JobName      string
	LastRun      time.Time
	LastExitCode int
	RunCount     int
	FailCount    int
	DriftCount   int
}

// History maintains a thread-safe map of job entries.
type History struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New returns an initialised History.
func New() *History {
	return &History{
		entries: make(map[string]*Entry),
	}
}

// Record updates the history for the named job.
func (h *History) Record(jobName string, exitCode int, drifted bool, at time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()

	e, ok := h.entries[jobName]
	if !ok {
		e = &Entry{JobName: jobName}
		h.entries[jobName] = e
	}

	e.LastRun = at
	e.LastExitCode = exitCode
	e.RunCount++
	if exitCode != 0 {
		e.FailCount++
	}
	if drifted {
		e.DriftCount++
	}
}

// Get returns a copy of the Entry for jobName and whether it was found.
func (h *History) Get(jobName string) (Entry, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	e, ok := h.entries[jobName]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all entries.
func (h *History) All() []Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]Entry, 0, len(h.entries))
	for _, e := range h.entries {
		out = append(out, *e)
	}
	return out
}

// Reset removes all history for jobName.
func (h *History) Reset(jobName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.entries, jobName)
}
