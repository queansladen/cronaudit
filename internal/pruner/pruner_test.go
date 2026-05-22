package pruner_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/pruner"
)

func makeFile(t *testing.T, path string, age time.Duration) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	mtime := time.Now().Add(-age)
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}

func TestPruneJob_RemovesOldFiles(t *testing.T) {
	base := t.TempDir()
	makeFile(t, filepath.Join(base, "myjob", "old.json"), 48*time.Hour)
	makeFile(t, filepath.Join(base, "myjob", "new.json"), 1*time.Hour)

	p := pruner.New(base, 24*time.Hour, nil)
	n, err := p.PruneJob("myjob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 file removed, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(base, "myjob", "new.json")); err != nil {
		t.Error("new.json should still exist")
	}
	if _, err := os.Stat(filepath.Join(base, "myjob", "old.json")); !os.IsNotExist(err) {
		t.Error("old.json should have been removed")
	}
}

func TestPruneJob_NonExistentDirIsNoop(t *testing.T) {
	base := t.TempDir()
	p := pruner.New(base, 24*time.Hour, nil)
	n, err := p.PruneJob("ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 removed, got %d", n)
	}
}

func TestPruneAll_PrunesMultipleJobs(t *testing.T) {
	base := t.TempDir()
	makeFile(t, filepath.Join(base, "job1", "stale.json"), 72*time.Hour)
	makeFile(t, filepath.Join(base, "job2", "stale.json"), 72*time.Hour)
	makeFile(t, filepath.Join(base, "job2", "fresh.json"), 2*time.Hour)

	p := pruner.New(base, 24*time.Hour, nil)
	n, err := p.PruneAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 removed, got %d", n)
	}
}

func TestPruneAll_EmptyBaseIsNoop(t *testing.T) {
	base := t.TempDir()
	p := pruner.New(base, 24*time.Hour, nil)
	n, err := p.PruneAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 removed, got %d", n)
	}
}

func TestPruneJob_KeepsAllWhenNothingExpired(t *testing.T) {
	base := t.TempDir()
	makeFile(t, filepath.Join(base, "myjob", "a.json"), 1*time.Minute)
	makeFile(t, filepath.Join(base, "myjob", "b.json"), 5*time.Minute)

	p := pruner.New(base, 24*time.Hour, nil)
	n, err := p.PruneJob("myjob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 removed, got %d", n)
	}
}
