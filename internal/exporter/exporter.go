// Package exporter writes cronaudit run results to JSON Lines format
// for consumption by external tools (e.g. log shippers, dashboards).
package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourorg/cronaudit/internal/store"
)

// Record is the JSON Lines schema emitted per run.
type Record struct {
	Timestamp time.Time `json:"timestamp"`
	Job       string    `json:"job"`
	ExitCode  int       `json:"exit_code"`
	Duration  float64   `json:"duration_seconds"`
	Drifted   bool      `json:"drifted"`
	Output    string    `json:"output,omitempty"`
}

// Exporter writes JSONL records to a destination writer.
type Exporter struct {
	w io.Writer
}

// New returns an Exporter that writes to w.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Exporter {
	if w == nil {
		w = os.Stdout
	}
	return &Exporter{w: w}
}

// Export converts a store.Result into a JSONL record and writes it.
func (e *Exporter) Export(jobName string, result *store.Result, drifted bool) error {
	if result == nil {
		return fmt.Errorf("exporter: result must not be nil")
	}
	if jobName == "" {
		return fmt.Errorf("exporter: jobName must not be empty")
	}
	rec := Record{
		Timestamp: result.RunAt,
		Job:       jobName,
		ExitCode:  result.ExitCode,
		Duration:  result.Duration.Seconds(),
		Drifted:   drifted,
		Output:    result.Output,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("exporter: marshal: %w", err)
	}
	_, err = fmt.Fprintf(e.w, "%s\n", b)
	if err != nil {
		return fmt.Errorf("exporter: write: %w", err)
	}
	return nil
}

// ExportToFile opens (or creates/appends) the file at path and exports the record.
func ExportToFile(path, jobName string, result *store.Result, drifted bool) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("exporter: open %s: %w", path, err)
	}
	defer f.Close()
	return New(f).Export(jobName, result, drifted)
}
