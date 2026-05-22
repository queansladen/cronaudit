// Package summarizer aggregates cron job run results into per-job statistics
// including total runs, failure counts, drift counts, average duration, and
// the timestamp of the most recent execution.
//
// It is intended to be used alongside the reporter and exporter packages to
// provide operators with a quick health overview of all monitored jobs.
package summarizer
