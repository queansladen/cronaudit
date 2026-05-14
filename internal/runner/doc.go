// Package runner provides functionality for executing shell commands
// associated with cronaudit jobs. It captures stdout, stderr, exit codes,
// and timing information, and enforces per-job or global timeouts.
//
// Basic usage:
//
//	r := runner.New(30 * time.Second)
//	result := r.Run(ctx, "my-job", "df -h /", 0)
//	if result.Error != nil {
//		log.Printf("job %s failed: %v", result.JobName, result.Error)
//	}
package runner
