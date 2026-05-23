// Package jobstats tracks cumulative per-job execution statistics including
// run counts, failure counts, drift counts, and duration averages.
//
// A Tracker is safe for concurrent use. Callers should invoke Record after
// each job execution and query Get or All when generating reports.
package jobstats
