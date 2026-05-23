// Package jobstats tracks cumulative per-job execution statistics including
// run counts, failure counts, drift counts, and duration averages.
//
// A Tracker is safe for concurrent use. Callers should invoke Record after
// each job execution and query Get or All when generating reports.
//
// # Overview
//
// Statistics are keyed by job name and accumulated in memory for the lifetime
// of the Tracker. Each call to Record updates the following fields:
//
//   - Runs: total number of executions
//   - Failures: number of executions that returned a non-nil error
//   - Drifts: number of executions that started later than their scheduled time
//   - AvgDuration: rolling average of execution wall-clock duration
//
// # Example
//
//	tracker := jobstats.NewTracker()
//	tracker.Record("backup", result)
//	stats, ok := tracker.Get("backup")
package jobstats
