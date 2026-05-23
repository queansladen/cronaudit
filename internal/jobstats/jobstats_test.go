package jobstats_test

import (
	"testing"
	"time"

	"github.com/cronaudit/cronaudit/internal/jobstats"
)

func newTracker() *jobstats.Tracker {
	return jobstats.New()
}

func TestRecord_InitialisesOnFirstCall(t *testing.T) {
	tr := newTracker()
	tr.Record("backup", 2*time.Second, false, false)
	s, ok := tr.Get("backup")
	if !ok {
		t.Fatal("expected stats to exist after first Record")
	}
	if s.TotalRuns != 1 {
		t.Fatalf("expected TotalRuns=1, got %d", s.TotalRuns)
	}
	if s.Failures != 0 {
		t.Fatalf("expected Failures=0, got %d", s.Failures)
	}
}

func TestRecord_IncrementsFailures(t *testing.T) {
	tr := newTracker()
	tr.Record("backup", time.Second, true, false)
	tr.Record("backup", time.Second, false, false)
	s, _ := tr.Get("backup")
	if s.Failures != 1 {
		t.Fatalf("expected Failures=1, got %d", s.Failures)
	}
}

func TestRecord_IncrementsDrifts(t *testing.T) {
	tr := newTracker()
	tr.Record("sync", time.Second, false, true)
	tr.Record("sync", time.Second, false, true)
	s, _ := tr.Get("sync")
	if s.Drifts != 2 {
		t.Fatalf("expected Drifts=2, got %d", s.Drifts)
	}
}

func TestAvgDuration_CorrectMean(t *testing.T) {
	tr := newTracker()
	tr.Record("job", 2*time.Second, false, false)
	tr.Record("job", 4*time.Second, false, false)
	s, _ := tr.Get("job")
	if s.AvgDuration() != 3*time.Second {
		t.Fatalf("expected avg 3s, got %s", s.AvgDuration())
	}
}

func TestAvgDuration_ZeroWhenNoRuns(t *testing.T) {
	s := jobstats.Stats{}
	if s.AvgDuration() != 0 {
		t.Fatal("expected 0 duration for empty stats")
	}
}

func TestSuccessRate_AllSuccess(t *testing.T) {
	tr := newTracker()
	tr.Record("job", time.Second, false, false)
	tr.Record("job", time.Second, false, false)
	s, _ := tr.Get("job")
	if s.SuccessRate() != 1.0 {
		t.Fatalf("expected success rate 1.0, got %f", s.SuccessRate())
	}
}

func TestSuccessRate_PartialFailures(t *testing.T) {
	tr := newTracker()
	tr.Record("job", time.Second, true, false)
	tr.Record("job", time.Second, false, false)
	tr.Record("job", time.Second, false, false)
	tr.Record("job", time.Second, false, false)
	s, _ := tr.Get("job")
	if got := s.SuccessRate(); got != 0.75 {
		t.Fatalf("expected 0.75, got %f", got)
	}
}

func TestGet_MissingJobReturnsFalse(t *testing.T) {
	tr := newTracker()
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Fatal("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	tr := newTracker()
	tr.Record("a", time.Second, false, false)
	tr.Record("b", time.Second, false, false)
	all := tr.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
