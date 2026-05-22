// Package pruner implements retention-based cleanup of cronaudit run result files.
//
// A Pruner is configured with a base storage directory and a retention duration.
// Any result file whose modification time is older than the retention window is
// deleted, keeping disk usage bounded over long-running deployments.
//
// Typical usage:
//
//	p := pruner.New("/var/lib/cronaudit", 7*24*time.Hour, os.Stderr)
//	if _, err := p.PruneAll(); err != nil {
//		log.Printf("pruning failed: %v", err)
//	}
package pruner
