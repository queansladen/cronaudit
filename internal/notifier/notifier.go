// Package notifier handles alerting when drift is detected in cron job output.
package notifier

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/example/cronaudit/internal/differ"
)

// Level represents the severity of a notification.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelDrift Level = "DRIFT"
)

// Event holds the details of a drift notification.
type Event struct {
	JobName   string
	Level     Level
	Message   string
	Diff      []differ.Change
	Timestamp time.Time
}

// Notifier writes drift events to a configured output.
type Notifier struct {
	out io.Writer
}

// New creates a Notifier that writes to the given writer.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Notifier {
	if w == nil {
		w = os.Stdout
	}
	return &Notifier{out: w}
}

// Notify formats and writes an Event to the output writer.
func (n *Notifier) Notify(e Event) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s | job=%s\n",
		e.Timestamp.UTC().Format(time.RFC3339),
		e.Level,
		e.JobName,
	))
	if e.Message != "" {
		sb.WriteString(fmt.Sprintf("  message: %s\n", e.Message))
	}
	for _, c := range e.Diff {
		switch c.Op {
		case differ.OpAdd:
			sb.WriteString(fmt.Sprintf("  + %s\n", c.Line))
		case differ.OpRemove:
			sb.WriteString(fmt.Sprintf("  - %s\n", c.Line))
		}
	}
	_, err := fmt.Fprint(n.out, sb.String())
	return err
}

// NotifyDrift is a convenience wrapper for emitting a DRIFT-level event.
func (n *Notifier) NotifyDrift(jobName string, changes []differ.Change) error {
	return n.Notify(Event{
		JobName:   jobName,
		Level:     LevelDrift,
		Message:   fmt.Sprintf("%d line(s) changed", len(changes)),
		Diff:      changes,
		Timestamp: time.Now(),
	})
}
