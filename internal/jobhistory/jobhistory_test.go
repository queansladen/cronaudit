package jobhistory_test

import (
	"testing"
	"time"

	"github.com/cronaudit/cronaudit/internal/jobhistory"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestRecord_FirstEntry(t *testing.T) {
	h := jobhistory.New()
	h.Record("backup", 0, false, epoch)

	e, ok := h.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.RunCount != 1 {
		t.Errorf("RunCount: got %d, want 1", e.RunCount)
	}
	if e.FailCount != 0 {
		t.Errorf("FailCount: got %d, want 0", e.FailCount)
	}
	if e.DriftCount != 0 {
		t.Errorf("DriftCount: got %d, want 0", e.DriftCount)
	}
}

func TestRecord_FailureIncrementsFailCount(t *testing.T) {
	h := jobhistory.New()
	h.Record("sync", 1, false, epoch)
	h.Record("sync", 0, false, epoch.Add(time.Minute))

	e, _ := h.Get("sync")
	if e.RunCount != 2 {
		t.Errorf("RunCount: got %d, want 2", e.RunCount)
	}
	if e.FailCount != 1 {
		t.Errorf("FailCount: got %d, want 1", e.FailCount)
	}
}

func TestRecord_DriftIncrement(t *testing.T) {
	h := jobhistory.New()
	h.Record("check", 0, true, epoch)
	h.Record("check", 0, true, epoch.Add(time.Hour))
	h.Record("check", 0, false, epoch.Add(2*time.Hour))

	e, _ := h.Get("check")
	if e.DriftCount != 2 {
		t.Errorf("DriftCount: got %d, want 2", e.DriftCount)
	}
}

func TestGet_MissingJobReturnsFalse(t *testing.T) {
	h := jobhistory.New()
	_, ok := h.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	h := jobhistory.New()
	h.Record("a", 0, false, epoch)
	h.Record("b", 0, false, epoch)
	h.Record("c", 1, true, epoch)

	all := h.All()
	if len(all) != 3 {
		t.Errorf("All: got %d entries, want 3", len(all))
	}
}

func TestReset_RemovesEntry(t *testing.T) {
	h := jobhistory.New()
	h.Record("temp", 0, false, epoch)
	h.Reset("temp")

	_, ok := h.Get("temp")
	if ok {
		t.Error("expected entry to be removed after Reset")
	}
}

func TestRecord_LastRunUpdated(t *testing.T) {
	h := jobhistory.New()
	later := epoch.Add(5 * time.Minute)
	h.Record("job", 0, false, epoch)
	h.Record("job", 0, false, later)

	e, _ := h.Get("job")
	if !e.LastRun.Equal(later) {
		t.Errorf("LastRun: got %v, want %v", e.LastRun, later)
	}
}
