package jobmeta

import (
	"testing"
	"time"
)

func TestRecord_InitialisesMetaOnFirstCall(t *testing.T) {
	tr := New()
	tr.Record("backup", 0, 2*time.Second, false)

	m, ok := tr.Get("backup")
	if !ok {
		t.Fatal("expected meta to exist after Record")
	}
	if m.RunCount != 1 {
		t.Errorf("RunCount = %d, want 1", m.RunCount)
	}
	if m.FailCount != 0 {
		t.Errorf("FailCount = %d, want 0", m.FailCount)
	}
}

func TestRecord_IncrementsFailCountOnNonZeroExit(t *testing.T) {
	tr := New()
	tr.Record("job", 1, time.Second, false)
	tr.Record("job", 0, time.Second, false)

	m, _ := tr.Get("job")
	if m.FailCount != 1 {
		t.Errorf("FailCount = %d, want 1", m.FailCount)
	}
	if m.RunCount != 2 {
		t.Errorf("RunCount = %d, want 2", m.RunCount)
	}
}

func TestRecord_IncrementsDriftCount(t *testing.T) {
	tr := New()
	tr.Record("job", 0, time.Second, true)
	tr.Record("job", 0, time.Second, false)
	tr.Record("job", 0, time.Second, true)

	m, _ := tr.Get("job")
	if m.DriftCount != 2 {
		t.Errorf("DriftCount = %d, want 2", m.DriftCount)
	}
}

func TestAvgDuration_ReturnsCorrectMean(t *testing.T) {
	tr := New()
	tr.Record("job", 0, 4*time.Second, false)
	tr.Record("job", 0, 2*time.Second, false)

	m, _ := tr.Get("job")
	if got := m.AvgDuration(); got != 3*time.Second {
		t.Errorf("AvgDuration = %v, want 3s", got)
	}
}

func TestAvgDuration_ZeroWhenNoRuns(t *testing.T) {
	m := Meta{}
	if m.AvgDuration() != 0 {
		t.Error("expected zero duration for empty meta")
	}
}

func TestGet_MissingJobReturnsFalse(t *testing.T) {
	tr := New()
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsAllTrackedJobs(t *testing.T) {
	tr := New()
	tr.Record("alpha", 0, time.Second, false)
	tr.Record("beta", 0, time.Second, false)
	tr.Record("alpha", 0, time.Second, false)

	all := tr.All()
	if len(all) != 2 {
		t.Errorf("All returned %d entries, want 2", len(all))
	}
}

func TestRecord_LastExitCodeUpdated(t *testing.T) {
	tr := New()
	tr.Record("job", 1, time.Second, false)
	tr.Record("job", 0, time.Second, false)

	m, _ := tr.Get("job")
	if m.LastExitCode != 0 {
		t.Errorf("LastExitCode = %d, want 0", m.LastExitCode)
	}
}
