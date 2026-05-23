// Package jobpolicy provides per-job execution policy evaluation for cronaudit.
//
// A Policy constrains when a job may run by specifying:
//   - allowed weekdays
//   - a time-of-day window (UTC)
//   - a maximum consecutive-failure threshold
//
// The Evaluator is safe for use within a single goroutine. Callers that share
// an Evaluator across goroutines must synchronise access externally.
package jobpolicy
