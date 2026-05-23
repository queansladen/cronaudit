// Package notifier provides drift notification functionality for cronaudit.
//
// A Notifier receives drift events — produced when cron job output differs
// from a previously recorded baseline — and writes formatted reports to a
// configured io.Writer (defaulting to stdout).
//
// Drift reports include the job name, a timestamp, and a structured summary
// of the detected changes (added lines, removed lines, or both). Each report
// is written atomically to the configured writer.
//
// Typical usage:
//
//	n := notifier.New(os.Stdout)
//	n.NotifyDrift("backup-job", changes)
//
// To suppress output during testing, pass io.Discard as the writer:
//
//	n := notifier.New(io.Discard)
package notifier
