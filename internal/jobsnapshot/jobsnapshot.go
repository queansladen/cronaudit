// Package jobsnapshot provides point-in-time snapshots of job execution state,
// allowing callers to capture and restore a consistent view of job metadata
// for reporting or diagnostics.
package jobsnapshot

import (
	"sync"
	"time"
)

// Snapshot holds an immutable view of a single job's state at capture time.
type Snapshot struct {
	JobName    string
	CapturedAt time.Time
	LastOutput string
	ExitCode   int
	Drifted    bool
	RunCount   int
	FailCount  int
}

// Manager captures and stores snapshots keyed by job name.
type Manager struct {
	mu        sync.RWMutex
	snapshots map[string]Snapshot
}

// New returns an initialised Manager.
func New() *Manager {
	return &Manager{
		snapshots: make(map[string]Snapshot),
	}
}

// Capture records a snapshot for the given job, overwriting any previous entry.
func (m *Manager) Capture(s Snapshot) {
	if s.JobName == "" {
		return
	}
	if s.CapturedAt.IsZero() {
		s.CapturedAt = time.Now()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[s.JobName] = s
}

// Get returns the most recent snapshot for a job and whether it was found.
func (m *Manager) Get(jobName string) (Snapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.snapshots[jobName]
	return s, ok
}

// All returns a copy of every stored snapshot.
func (m *Manager) All() []Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Snapshot, 0, len(m.snapshots))
	for _, s := range m.snapshots {
		out = append(out, s)
	}
	return out
}

// Delete removes the snapshot for a job, if present.
func (m *Manager) Delete(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.snapshots, jobName)
}

// Reset removes all stored snapshots.
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots = make(map[string]Snapshot)
}
