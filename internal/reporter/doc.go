// Package reporter aggregates cron job run results from the store and
// produces human-readable drift summary reports over a configurable time
// window.
//
// Usage:
//
//	r := reporter.New(store, os.Stdout)
//	rpt, err := r.Generate(jobNames, from, to)
//	if err != nil { ... }
//	r.Print(rpt)
package reporter
