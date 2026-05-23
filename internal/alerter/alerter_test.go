package alerter

import (
	"bytes"
	"strings"
	"testing"
)

func newTestAlerter(threshold int) (*Alerter, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return New(threshold, buf), buf
}

func TestRecord_NoDriftResetsCounter(t *testing.T) {
	a, _ := newTestAlerter(2)
	a.Record("myjob", true)
	a.Record("myjob", false)

	a.mu.Lock()
	count := a.counts["myjob"]
	a.mu.Unlock()

	if count != 0 {
		t.Fatalf("expected counter to reset to 0, got %d", count)
	}
}

func TestRecord_BelowThresholdReturnsNil(t *testing.T) {
	a, buf := newTestAlerter(3)
	alert := a.Record("myjob", true)
	if alert != nil {
		t.Fatal("expected nil alert below threshold")
	}
	if buf.Len() != 0 {
		t.Fatal("expected no output below threshold")
	}
}

func TestRecord_ThresholdReachedReturnsAlert(t *testing.T) {
	a, buf := newTestAlerter(2)
	a.Record("myjob", true)
	alert := a.Record("myjob", true)

	if alert == nil {
		t.Fatal("expected non-nil alert at threshold")
	}
	if alert.JobName != "myjob" {
		t.Errorf("expected job name 'myjob', got %q", alert.JobName)
	}
	if alert.ConsecutiveDrift != 2 {
		t.Errorf("expected ConsecutiveDrift=2, got %d", alert.ConsecutiveDrift)
	}
	if alert.Threshold != 2 {
		t.Errorf("expected Threshold=2, got %d", alert.Threshold)
	}
	if !strings.Contains(buf.String(), "myjob") {
		t.Errorf("expected alert output to mention job name, got: %s", buf.String())
	}
}

func TestRecord_ExceedingThresholdContinuesToAlert(t *testing.T) {
	a, _ := newTestAlerter(2)
	a.Record("myjob", true)
	a.Record("myjob", true)
	alert := a.Record("myjob", true)

	if alert == nil {
		t.Fatal("expected alert beyond threshold")
	}
	if alert.ConsecutiveDrift != 3 {
		t.Errorf("expected ConsecutiveDrift=3, got %d", alert.ConsecutiveDrift)
	}
}

func TestReset_ClearsCounter(t *testing.T) {
	a, _ := newTestAlerter(5)
	a.Record("myjob", true)
	a.Record("myjob", true)
	a.Reset("myjob")

	a.mu.Lock()
	count := a.counts["myjob"]
	a.mu.Unlock()

	if count != 0 {
		t.Fatalf("expected counter 0 after Reset, got %d", count)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	a := New(1, nil)
	if a.out == nil {
		t.Fatal("expected non-nil writer when nil passed to New")
	}
}

func TestNew_ThresholdMinimumIsOne(t *testing.T) {
	a, _ := newTestAlerter(0)
	if a.threshold != 1 {
		t.Errorf("expected threshold clamped to 1, got %d", a.threshold)
	}
}
