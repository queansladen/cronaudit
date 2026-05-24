// Package ratelimiter provides a simple token-bucket rate limiter
// for controlling how frequently drift alerts are emitted per job.
package ratelimiter

import (
	"sync"
	"time"
)

// Limiter tracks per-job alert rate limiting using a sliding window.
type Limiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxCalls int
	history  map[string][]time.Time
	now      func() time.Time
}

// New creates a Limiter that allows at most maxCalls events per job
// within the given window duration.
func New(window time.Duration, maxCalls int) *Limiter {
	return &Limiter{
		window:   window,
		maxCalls: maxCalls,
		history:  make(map[string][]time.Time),
		now:      time.Now,
	}
}

// Allow reports whether the named job is permitted to emit an alert
// right now. It records the attempt if allowed.
func (l *Limiter) Allow(job string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	// Prune timestamps outside the window.
	times := l.history[job]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.maxCalls {
		l.history[job] = valid
		return false
	}

	l.history[job] = append(valid, now)
	return true
}

// Reset clears the history for a specific job.
func (l *Limiter) Reset(job string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.history, job)
}

// Remaining returns how many more calls are allowed for the job within
// the current window.
func (l *Limiter) Remaining(job string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window)

	count := 0
	for _, t := range l.history[job] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := l.maxCalls - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ResetAll clears the rate-limit history for all tracked jobs.
// This is useful when reconfiguring limits or during testing teardown.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.history = make(map[string][]time.Time)
}
