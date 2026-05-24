package jobpause_test

import (
	"testing"
	"time"

	"github.com/example/cronaudit/internal/jobpause"
)

func newManager() *jobpause.Manager {
	return jobpause.New()
}

func TestIsPaused_NotPausedByDefault(t *testing.T) {
	m := newManager()
	if m.IsPaused("backup") {
		t.Fatal("expected job to not be paused")
	}
}

func TestPause_JobIsPaused(t *testing.T) {
	m := newManager()
	m.Pause("backup", time.Now().Add(time.Hour), "maintenance")
	if !m.IsPaused("backup") {
		t.Fatal("expected job to be paused")
	}
}

func TestPause_ExpiredPauseIsNotPaused(t *testing.T) {
	m := newManager()
	m.Pause("backup", time.Now().Add(-time.Second), "old")
	if m.IsPaused("backup") {
		t.Fatal("expired pause should not mark job as paused")
	}
}

func TestPause_ZeroTimeIsNoop(t *testing.T) {
	m := newManager()
	m.Pause("backup", time.Time{}, "")
	if m.IsPaused("backup") {
		t.Fatal("zero-time pause should be a no-op")
	}
}

func TestResume_ClearsPause(t *testing.T) {
	m := newManager()
	m.Pause("backup", time.Now().Add(time.Hour), "test")
	m.Resume("backup")
	if m.IsPaused("backup") {
		t.Fatal("expected pause to be cleared after Resume")
	}
}

func TestGet_ReturnsPauseRecord(t *testing.T) {
	m := newManager()
	until := time.Now().Add(time.Hour)
	m.Pause("sync", until, "deploy")
	p, ok := m.Get("sync")
	if !ok {
		t.Fatal("expected pause record to exist")
	}
	if p.Reason != "deploy" {
		t.Fatalf("expected reason 'deploy', got %q", p.Reason)
	}
	if !p.Until.Equal(until) {
		t.Fatalf("unexpected Until value")
	}
}

func TestGet_MissingJobReturnsFalse(t *testing.T) {
	m := newManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestAll_ReturnsOnlyActivePauses(t *testing.T) {
	m := newManager()
	m.Pause("active", time.Now().Add(time.Hour), "reason")
	m.Pause("expired", time.Now().Add(-time.Second), "old")
	all := m.All()
	if _, ok := all["active"]; !ok {
		t.Fatal("expected active job in All()")
	}
	if _, ok := all["expired"]; ok {
		t.Fatal("expired job should not appear in All()")
	}
}

func TestAll_IndependentJobs(t *testing.T) {
	m := newManager()
	m.Pause("job1", time.Now().Add(time.Hour), "")
	m.Pause("job2", time.Now().Add(2*time.Hour), "")
	all := m.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 paused jobs, got %d", len(all))
	}
}
