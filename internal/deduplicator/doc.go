// Package deduplicator provides output-level deduplication for cron job runs.
//
// When a cron job produces identical output on consecutive executions,
// the deduplicator suppresses downstream processing (notifications, exports)
// to reduce noise. Each job is tracked independently by name using a SHA-256
// hash of its stdout/stderr output.
package deduplicator
