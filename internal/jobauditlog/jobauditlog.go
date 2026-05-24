// Package jobauditlog maintains an append-only in-memory audit trail of
// significant job lifecycle events (start, success, failure, drift).
package jobauditlog

import (
	"sync"
	"time"
)

// EventKind classifies a single audit event.
type EventKind string

const (
	EventStart   EventKind = "start"
	EventSuccess EventKind = "success"
	EventFailure EventKind = "failure"
	EventDrift   EventKind = "drift"
)

// Entry is one record in the audit log.
type Entry struct {
	JobName   string
	Kind      EventKind
	Timestamp time.Time
	Detail    string // optional free-text (e.g. diff summary, error message)
}

// Log holds the audit trail for all jobs.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	maxSize int
}

// New creates a Log that retains at most maxSize entries.
// If maxSize <= 0 the log is unbounded.
func New(maxSize int) *Log {
	return &Log{maxSize: maxSize}
}

// Record appends a new entry to the log.
func (l *Log) Record(job string, kind EventKind, detail string) {
	if job == "" {
		return
	}
	e := Entry{
		JobName:   job,
		Kind:      kind,
		Timestamp: time.Now().UTC(),
		Detail:    detail,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, e)
	if l.maxSize > 0 && len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

// All returns a shallow copy of all entries in insertion order.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// ForJob returns entries for the named job in insertion order.
func (l *Log) ForJob(job string) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var out []Entry
	for _, e := range l.entries {
		if e.JobName == job {
			out = append(out, e)
		}
	}
	return out
}

// Len returns the current number of stored entries.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
