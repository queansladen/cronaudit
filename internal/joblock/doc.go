// Package joblock prevents concurrent execution of the same cron job
// by maintaining a lightweight in-memory lock per job name.
//
// When a scheduler tick fires for a job that is still running from a
// previous tick, Acquire returns an error so the caller can skip that
// invocation rather than stacking duplicate runs.
package joblock
