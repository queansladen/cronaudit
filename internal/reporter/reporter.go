// Package reporter generates periodic summary reports of cron job drift
// activity, aggregating results from the store over a configurable window.
package reporter

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/example/cronaudit/internal/store"
)

// Summary holds aggregated statistics for a single job over a time window.
type Summary struct {
	JobName    string
	Runs       int
	Drifts     int
	LastRun    time.Time
	LastOutput string
}

// Report holds all job summaries for a reporting period.
type Report struct {
	GeneratedAt time.Time
	WindowStart time.Time
	WindowEnd   time.Time
	Summaries   []Summary
}

// Reporter generates drift summary reports from stored run results.
type Reporter struct {
	w     io.Writer
	store *store.Store
}

// New creates a Reporter writing to w. If w is nil, os.Stdout is used.
func New(s *store.Store, w io.Writer) *Reporter {
	if w == nil {
		w = os.Stdout
	}
	return &Reporter{w: w, store: s}
}

// Generate builds a Report for the given job names over [from, to].
func (r *Reporter) Generate(jobNames []string, from, to time.Time) (*Report, error) {
	rpt := &Report{
		GeneratedAt: time.Now(),
		WindowStart: from,
		WindowEnd:   to,
	}

	for _, name := range jobNames {
		results, err := r.store.Range(name, from, to)
		if err != nil {
			return nil, fmt.Errorf("reporter: fetching %q: %w", name, err)
		}

		sum := Summary{JobName: name, Runs: len(results)}
		for _, res := range results {
			if res.Drifted {
				sum.Drifts++
			}
			if res.RunAt.After(sum.LastRun) {
				sum.LastRun = res.RunAt
				sum.LastOutput = res.Output
			}
		}
		rpt.Summaries = append(rpt.Summaries, sum)
	}
	return rpt, nil
}

// Print writes a human-readable summary of the report to the reporter's writer.
func (r *Reporter) Print(rpt *Report) {
	fmt.Fprintf(r.w, "CronAudit Report — generated %s\n", rpt.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(r.w, "Window: %s → %s\n\n", rpt.WindowStart.Format(time.RFC3339), rpt.WindowEnd.Format(time.RFC3339))

	for _, s := range rpt.Summaries {
		driftPct := 0.0
		if s.Runs > 0 {
			driftPct = float64(s.Drifts) / float64(s.Runs) * 100
		}
		fmt.Fprintf(r.w, "  %-30s runs=%-4d drifts=%-4d (%.1f%%)\n",
			s.JobName, s.Runs, s.Drifts, driftPct)
	}
}
