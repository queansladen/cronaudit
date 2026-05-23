// Package circuitbreaker provides a simple circuit breaker for cron job
// execution, preventing repeated runs of consistently failing jobs.
package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State represents the current state of a circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking execution
	StateHalfOpen              // testing recovery
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Breaker tracks failure counts per job and opens when the threshold is exceeded.
type Breaker struct {
	mu           sync.Mutex
	failures     map[string]int
	states       map[string]State
	openedAt     map[string]time.Time
	threshold    int
	recoveryWait time.Duration
}

// New creates a Breaker that opens after threshold consecutive failures
// and attempts recovery after recoveryWait.
func New(threshold int, recoveryWait time.Duration) *Breaker {
	return &Breaker{
		failures:     make(map[string]int),
		states:       make(map[string]State),
		openedAt:     make(map[string]time.Time),
		threshold:    threshold,
		recoveryWait: recoveryWait,
	}
}

// Allow returns nil if the job is permitted to run, or an error if the
// circuit is open.
func (b *Breaker) Allow(job string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.states[job]

	if state == StateOpen {
		if time.Since(b.openedAt[job]) >= b.recoveryWait {
			b.states[job] = StateHalfOpen
			return nil
		}
		return fmt.Errorf("circuit open for job %q: too many consecutive failures", job)
	}

	return nil
}

// RecordSuccess resets the failure counter and closes the circuit for job.
func (b *Breaker) RecordSuccess(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures[job] = 0
	b.states[job] = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit if
// the threshold is reached.
func (b *Breaker) RecordFailure(job string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures[job]++
	if b.failures[job] >= b.threshold {
		b.states[job] = StateOpen
		b.openedAt[job] = time.Now()
	}
}

// State returns the current circuit state for a job.
func (b *Breaker) State(job string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.states[job]
}
