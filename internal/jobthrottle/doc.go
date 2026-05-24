// Package jobthrottle provides per-job concurrency limiting for cronaudit.
//
// A Throttle enforces a maximum number of simultaneous executions for any
// named job. Callers must call Acquire before starting a job and Release
// (typically via defer) when the job finishes.
//
// Example usage:
//
//	th := jobthrottle.New(2)
//	if err := th.Acquire(job.Name); err != nil {
//		// job is throttled – skip or queue
//		return err
//	}
//	defer th.Release(job.Name)
//	// … run the job …
package jobthrottle
