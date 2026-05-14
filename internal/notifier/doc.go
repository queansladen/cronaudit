// Package notifier provides drift notification functionality for cronaudit.
//
// A Notifier receives drift events — produced when cron job output differs
// from a previously recorded baseline — and writes formatted reports to a
// configured io.Writer (defaulting to stdout).
//
// Typical usage:
//
//	n := notifier.New(os.Stdout)
//	n.NotifyDrift("backup-job", changes)
package notifier
