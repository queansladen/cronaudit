// Package ratelimiter implements a sliding-window rate limiter used by
// cronaudit to suppress repeated drift alerts for the same job within a
// configurable time window.
//
// Usage:
//
//	l := ratelimiter.New(5*time.Minute, 3)
//	if l.Allow(jobName) {
//		// emit alert
//	}
package ratelimiter
