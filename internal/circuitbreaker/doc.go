// Package circuitbreaker implements a per-job circuit breaker for cronaudit.
//
// When a job fails consecutively beyond a configured threshold the circuit
// opens and subsequent Allow calls return an error, preventing the job from
// running until the recovery window has elapsed.  After the window the
// circuit moves to half-open, allowing a single trial run.  A successful
// run closes the circuit; another failure re-opens it.
package circuitbreaker
