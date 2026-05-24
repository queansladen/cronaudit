package jobsnapshot_test

import (
	"testing"
	"time"

	"github.com/example/cronaudit/internal/jobsnapshot"
)

func baseSnapshot(name string) jobsnapshot.Snapshot {
	return jobsnapshot.Snapshot{
		JobName:    name,
		CapturedAt: time.Now(),
		LastOutput: "ok",
		ExitCode:   0,
		Drifted:    false,
		RunCount:   1,
		FailCount:  0,
	}
}

func TestCapture_StoresSnapshot(t *testing.T) {
	m := jobsnapshot.New()
	s := baseSnapshot("backup")
	m.Capture(s)
	got, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected snapshot to be found")
	}
	if got.JobName != "backup" {
		t.Errorf("got job name %q, want %q", got.JobName, "backup")
	}
}

func TestCapture_OverwritesPrevious(t *testing.T) {
	m := jobsnapshot.New()
	first := baseSnapshot("sync")
	first.RunCount = 1
	m.Capture(first)
	second := baseSnapshot("sync")
	second.RunCount = 5
	m.Capture(second)
	got, _ := m.Get("sync")
	if got.RunCount != 5 {
		t.Errorf("got RunCount %d, want 5", got.RunCount)
	}
}

func TestCapture_EmptyJobNameIsNoop(t *testing.T) {
	m := jobsnapshot.New()
	m.Capture(jobsnapshot.Snapshot{JobName: ""})
	if len(m.All()) != 0 {
		t.Error("expected no snapshots for empty job name")
	}
}

func TestCapture_SetsTimestampWhenZero(t *testing.T) {
	m := jobsnapshot.New()
	before := time.Now()
	m.Capture(jobsnapshot.Snapshot{JobName: "clean"})
	got, _ := m.Get("clean")
	if got.CapturedAt.Before(before) {
		t.Error("expected CapturedAt to be set automatically")
	}
}

func TestGet_MissingJobReturnsFalse(t *testing.T) {
	m := jobsnapshot.New()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for missing job")
	}
}

func TestAll_ReturnsAllSnapshots(t *testing.T) {
	m := jobsnapshot.New()
	m.Capture(baseSnapshot("a"))
	m.Capture(baseSnapshot("b"))
	m.Capture(baseSnapshot("c"))
	if len(m.All()) != 3 {
		t.Errorf("got %d snapshots, want 3", len(m.All()))
	}
}

func TestDelete_RemovesSnapshot(t *testing.T) {
	m := jobsnapshot.New()
	m.Capture(baseSnapshot("tmp"))
	m.Delete("tmp")
	_, ok := m.Get("tmp")
	if ok {
		t.Error("expected snapshot to be deleted")
	}
}

func TestReset_ClearsAll(t *testing.T) {
	m := jobsnapshot.New()
	m.Capture(baseSnapshot("x"))
	m.Capture(baseSnapshot("y"))
	m.Reset()
	if len(m.All()) != 0 {
		t.Errorf("expected 0 snapshots after reset, got %d", len(m.All()))
	}
}
