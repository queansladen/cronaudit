// Package jobdispatcher provides a priority-aware worker pool that dequeues
// scheduled cron jobs and executes them concurrently while honouring per-job
// mutual-exclusion via joblock.
//
// Typical usage:
//
//	q, _ := jobqueue.New(128)
//	l := joblock.New()
//	d, _ := jobdispatcher.New(q, l, 4)
//	d.Start(ctx)
//	defer d.Stop()
//
// Errors from individual job runs are surfaced through Dispatcher.Errors().
package jobdispatcher
