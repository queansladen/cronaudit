// Package jobcooldown provides a per-job cooldown mechanism that enforces a
// minimum interval between successive executions. It is used by the scheduler
// to prevent thrashing when a job repeatedly fails or produces drift output.
package jobcooldown
