// Package summarizer produces human-readable run summaries for a set of
// cron job results, aggregating exit codes, durations, and drift counts.
package summarizer

import (
	"fmt"
	"io"
	"os"
	"time"
)

// JobSummary holds aggregated statistics for a single job.
type JobSummary struct {
	Name        string
	TotalRuns   int
	FailedRuns  int
	DriftCount  int
	AvgDuration time.Duration
	LastRun     time.Time
}

// Result is a minimal view of a stored run result used as input.
type Result struct {
	JobName  string
	ExitCode int
	Drifted  bool
	Duration time.Duration
	RunAt    time.Time
}

// Summarizer aggregates Results into per-job summaries.
type Summarizer struct {
	w io.Writer
}

// New returns a Summarizer that writes to w. If w is nil, os.Stdout is used.
func New(w io.Writer) *Summarizer {
	if w == nil {
		w = os.Stdout
	}
	return &Summarizer{w: w}
}

// Summarize computes per-job statistics from results and returns the map.
func (s *Summarizer) Summarize(results []Result) map[string]*JobSummary {
	summaries := make(map[string]*JobSummary)

	for _, r := range results {
		js, ok := summaries[r.JobName]
		if !ok {
			js = &JobSummary{Name: r.JobName}
			summaries[r.JobName] = js
		}
		js.TotalRuns++
		if r.ExitCode != 0 {
			js.FailedRuns++
		}
		if r.Drifted {
			js.DriftCount++
		}
		js.AvgDuration += r.Duration
		if r.RunAt.After(js.LastRun) {
			js.LastRun = r.RunAt
		}
	}

	for _, js := range summaries {
		if js.TotalRuns > 0 {
			js.AvgDuration /= time.Duration(js.TotalRuns)
		}
	}

	return summaries
}

// Print writes a formatted summary table to the configured writer.
func (s *Summarizer) Print(summaries map[string]*JobSummary) {
	fmt.Fprintf(s.w, "%-24s %8s %8s %8s %12s %s\n",
		"JOB", "RUNS", "FAILED", "DRIFTS", "AVG_DUR", "LAST_RUN")
	fmt.Fprintf(s.w, "%s\n", "------------------------------------------------------------------------")
	for _, js := range summaries {
		fmt.Fprintf(s.w, "%-24s %8d %8d %8d %12s %s\n",
			js.Name,
			js.TotalRuns,
			js.FailedRuns,
			js.DriftCount,
			js.AvgDuration.Round(time.Millisecond),
			js.LastRun.Format(time.RFC3339),
		)
	}
}
