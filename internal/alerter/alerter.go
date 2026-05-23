// Package alerter provides threshold-based alerting for consecutive drift events.
// When a job drifts more than a configured number of times in a row, an alert
// is emitted to notify operators of persistent output instability.
package alerter

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Alert represents a threshold breach for a specific job.
type Alert struct {
	JobName          string
	ConsecutiveDrift int
	Threshold        int
	At               time.Time
}

// Alerter tracks consecutive drift counts per job and emits alerts.
type Alerter struct {
	mu        sync.Mutex
	threshold int
	counts    map[string]int
	out       io.Writer
}

// New returns an Alerter that fires when a job drifts consecutively
// at least threshold times. Output defaults to os.Stdout.
func New(threshold int, out io.Writer) *Alerter {
	if out == nil {
		out = os.Stdout
	}
	if threshold < 1 {
		threshold = 1
	}
	return &Alerter{
		threshold: threshold,
		counts:    make(map[string]int),
		out:       out,
	}
}

// Record updates the consecutive drift counter for jobName.
// drifted=true increments the counter; false resets it.
// Returns a non-nil Alert when the threshold is reached or exceeded.
func (a *Alerter) Record(jobName string, drifted bool) *Alert {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !drifted {
		a.counts[jobName] = 0
		return nil
	}

	a.counts[jobName]++
	count := a.counts[jobName]

	if count >= a.threshold {
		alert := &Alert{
			JobName:          jobName,
			ConsecutiveDrift: count,
			Threshold:        a.threshold,
			At:               time.Now().UTC(),
		}
		a.emit(alert)
		return alert
	}
	return nil
}

// Reset clears the drift counter for jobName.
func (a *Alerter) Reset(jobName string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.counts[jobName] = 0
}

func (a *Alerter) emit(alert *Alert) {
	fmt.Fprintf(
		a.out,
		"[ALERT] %s — job %q has drifted %d consecutive time(s) (threshold: %d)\n",
		alert.At.Format(time.RFC3339),
		alert.JobName,
		alert.ConsecutiveDrift,
		alert.Threshold,
	)
}
