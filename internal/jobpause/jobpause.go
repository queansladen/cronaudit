// Package jobpause provides a mechanism to temporarily pause and resume
// individual cron jobs without removing them from the configuration.
package jobpause

import (
	"sync"
	"time"
)

// Pause records a pause window for a job.
type Pause struct {
	Until  time.Time
	Reason string
}

// Manager tracks paused jobs.
type Manager struct {
	mu     sync.RWMutex
	paused map[string]Pause
}

// New returns an initialised Manager.
func New() *Manager {
	return &Manager{
		paused: make(map[string]Pause),
	}
}

// Pause pauses job until the given time with an optional reason.
// Passing a zero-value time or a time in the past is a no-op.
func (m *Manager) Pause(job string, until time.Time, reason string) {
	if !until.After(time.Now()) {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paused[job] = Pause{Until: until, Reason: reason}
}

// Resume removes a pause for the given job, allowing it to run immediately.
func (m *Manager) Resume(job string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.paused, job)
}

// IsPaused reports whether the job is currently paused.
// Expired pauses are lazily evicted on access.
func (m *Manager) IsPaused(job string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.paused[job]
	if !ok {
		return false
	}
	if time.Now().After(p.Until) {
		delete(m.paused, job)
		return false
	}
	return true
}

// Get returns the Pause record for a job and whether one exists.
func (m *Manager) Get(job string) (Pause, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.paused[job]
	return p, ok
}

// All returns a snapshot of all currently active (non-expired) pauses.
func (m *Manager) All() map[string]Pause {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	out := make(map[string]Pause, len(m.paused))
	for job, p := range m.paused {
		if now.After(p.Until) {
			delete(m.paused, job)
			continue
		}
		out[job] = p
	}
	return out
}
