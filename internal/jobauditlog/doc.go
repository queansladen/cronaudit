// Package jobauditlog provides a thread-safe, append-only audit trail for
// cronaudit job lifecycle events.
//
// Events are classified as start, success, failure, or drift. The log can be
// bounded to a maximum number of entries; once the cap is exceeded the oldest
// entries are evicted so memory usage stays constant during long daemon runs.
//
// Typical usage:
//
//	log := jobauditlog.New(1000)
//	log.Record("backup", jobauditlog.EventStart, "")
//	log.Record("backup", jobauditlog.EventDrift, "+3 lines changed")
//	entries := log.ForJob("backup")
package jobauditlog
