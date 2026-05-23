// Package joblock provides a simple in-memory mutex-based lock
// to prevent concurrent execution of the same cron job.
package joblock

import (
	"fmt"
	"sync"
)

// Locker manages per-job execution locks.
type Locker struct {
	mu    sync.Mutex
	locks map[string]bool
}

// New returns a new Locker.
func New() *Locker {
	return &Locker{
		locks: make(map[string]bool),
	}
}

// Acquire attempts to acquire the lock for the given job name.
// It returns a release function and nil on success, or an error if the
// job is already running.
func (l *Locker) Acquire(jobName string) (release func(), err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.locks[jobName] {
		return nil, fmt.Errorf("joblock: job %q is already running", jobName)
	}

	l.locks[jobName] = true

	return func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.locks, jobName)
	}, nil
}

// IsLocked reports whether the given job currently holds a lock.
func (l *Locker) IsLocked(jobName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.locks[jobName]
}

// ActiveJobs returns the names of all currently locked jobs.
func (l *Locker) ActiveJobs() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	names := make([]string, 0, len(l.locks))
	for name := range l.locks {
		names = append(names, name)
	}
	return names
}
