package jobauditlog_test

import (
	"testing"

	"github.com/cronaudit/cronaudit/internal/jobauditlog"
)

func newLog(max int) *jobauditlog.Log {
	return jobauditlog.New(max)
}

func TestRecord_EmptyJobNameIsNoop(t *testing.T) {
	l := newLog(0)
	l.Record("", jobauditlog.EventStart, "")
	if l.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", l.Len())
	}
}

func TestRecord_AddsEntry(t *testing.T) {
	l := newLog(0)
	l.Record("myjob", jobauditlog.EventSuccess, "ok")
	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}
	all := l.All()
	if all[0].JobName != "myjob" || all[0].Kind != jobauditlog.EventSuccess {
		t.Errorf("unexpected entry: %+v", all[0])
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := newLog(0)
	l.Record("job", jobauditlog.EventStart, "")
	a := l.All()
	a[0].JobName = "mutated"
	b := l.All()
	if b[0].JobName == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestForJob_FiltersCorrectly(t *testing.T) {
	l := newLog(0)
	l.Record("alpha", jobauditlog.EventStart, "")
	l.Record("beta", jobauditlog.EventDrift, "diff")
	l.Record("alpha", jobauditlog.EventSuccess, "")

	alpha := l.ForJob("alpha")
	if len(alpha) != 2 {
		t.Fatalf("expected 2 entries for alpha, got %d", len(alpha))
	}
	for _, e := range alpha {
		if e.JobName != "alpha" {
			t.Errorf("unexpected job name: %s", e.JobName)
		}
	}
}

func TestMaxSize_EvictsOldestEntries(t *testing.T) {
	l := newLog(3)
	for i := 0; i < 5; i++ {
		l.Record("job", jobauditlog.EventSuccess, "")
	}
	if l.Len() != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", l.Len())
	}
}

func TestRecord_MultipleKinds(t *testing.T) {
	l := newLog(0)
	kinds := []jobauditlog.EventKind{
		jobauditlog.EventStart,
		jobauditlog.EventSuccess,
		jobauditlog.EventFailure,
		jobauditlog.EventDrift,
	}
	for _, k := range kinds {
		l.Record("job", k, "")
	}
	all := l.All()
	for i, e := range all {
		if e.Kind != kinds[i] {
			t.Errorf("index %d: expected %s got %s", i, kinds[i], e.Kind)
		}
	}
}

func TestTimestamp_IsSet(t *testing.T) {
	l := newLog(0)
	l.Record("job", jobauditlog.EventStart, "")
	e := l.All()[0]
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}
