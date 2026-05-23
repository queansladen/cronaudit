// Package jobpolicy evaluates per-job execution policies, determining
// whether a job should run based on its schedule window, allowed days,
// and maximum consecutive failure limits.
package jobpolicy

import (
	"errors"
	"fmt"
	"time"
)

// Policy defines the execution constraints for a single cron job.
type Policy struct {
	// JobName is the identifier of the job this policy applies to.
	JobName string

	// AllowedDays lists weekdays on which the job may run (0=Sunday … 6=Saturday).
	// Empty means all days are allowed.
	AllowedDays []time.Weekday

	// WindowStart and WindowEnd define an optional time-of-day window (UTC).
	// Both must be set together; zero values mean no window restriction.
	WindowStart time.Duration
	WindowEnd   time.Duration

	// MaxConsecutiveFailures is the maximum number of back-to-back failures
	// before the policy blocks further execution. 0 means no limit.
	MaxConsecutiveFailures int
}

// Evaluator checks whether a job is permitted to run under its policy.
type Evaluator struct {
	policies           map[string]Policy
	consecutiveFails   map[string]int
}

// New returns an Evaluator pre-loaded with the given policies.
func New(policies []Policy) *Evaluator {
	m := make(map[string]Policy, len(policies))
	for _, p := range policies {
		m[p.JobName] = p
	}
	return &Evaluator{
		policies:         m,
		consecutiveFails: make(map[string]int),
	}
}

// Allow returns nil if the job is permitted to run at t, or a descriptive
// error explaining which policy constraint blocks execution.
func (e *Evaluator) Allow(jobName string, t time.Time) error {
	p, ok := e.policies[jobName]
	if !ok {
		return nil // no policy → always allowed
	}

	if len(p.AllowedDays) > 0 {
		allowed := false
		for _, d := range p.AllowedDays {
			if t.Weekday() == d {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("job %q not allowed on %s", jobName, t.Weekday())
		}
	}

	if p.WindowStart != 0 || p.WindowEnd != 0 {
		tod := time.Duration(t.Hour())*time.Hour +
			time.Duration(t.Minute())*time.Minute +
			time.Duration(t.Second())*time.Second
		if tod < p.WindowStart || tod >= p.WindowEnd {
			return fmt.Errorf("job %q outside allowed window [%v, %v)", jobName, p.WindowStart, p.WindowEnd)
		}
	}

	if p.MaxConsecutiveFailures > 0 {
		if e.consecutiveFails[jobName] >= p.MaxConsecutiveFailures {
			return fmt.Errorf("job %q blocked after %d consecutive failures", jobName, e.consecutiveFails[jobName])
		}
	}

	return nil
}

// RecordResult updates the consecutive-failure counter for jobName.
// Pass exitCode == 0 for success.
func (e *Evaluator) RecordResult(jobName string, exitCode int) error {
	if _, ok := e.policies[jobName]; !ok {
		return errors.New("unknown job: " + jobName)
	}
	if exitCode == 0 {
		e.consecutiveFails[jobName] = 0
	} else {
		e.consecutiveFails[jobName]++
	}
	return nil
}

// ConsecutiveFailures returns the current consecutive-failure count for jobName.
func (e *Evaluator) ConsecutiveFailures(jobName string) int {
	return e.consecutiveFails[jobName]
}
