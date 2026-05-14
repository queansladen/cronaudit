package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/store"
)

func newTempStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

func TestSave_CreatesFile(t *testing.T) {
	s := newTempStore(t)
	rec := store.RunRecord{
		JobName:   "backup",
		Command:   "/usr/bin/backup.sh",
		StartedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Duration:  2 * time.Second,
		ExitCode:  0,
		Stdout:    "done\n",
		Stderr:    "",
	}
	if err := s.Save(rec); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func TestLatest_ReturnsNilWhenEmpty(t *testing.T) {
	s := newTempStore(t)
	rec, err := s.Latest("nonexistent")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if rec != nil {
		t.Errorf("expected nil, got %+v", rec)
	}
}

func TestLatest_ReturnsMostRecent(t *testing.T) {
	s := newTempStore(t)

	base := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	for i, stdout := range []string{"first run", "second run", "third run"} {
		err := s.Save(store.RunRecord{
			JobName:   "myjob",
			Command:   "echo hello",
			StartedAt: base.Add(time.Duration(i) * time.Hour),
			ExitCode:  0,
			Stdout:    stdout,
		})
		if err != nil {
			t.Fatalf("Save[%d]: %v", i, err)
		}
	}

	rec, err := s.Latest("myjob")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if rec == nil {
		t.Fatal("expected a record, got nil")
	}
	if rec.Stdout != "third run" {
		t.Errorf("expected %q, got %q", "third run", rec.Stdout)
	}
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := t.TempDir() + "/nested/path"
	_, err := store.New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("directory %q was not created", dir)
	}
}

func TestSave_SanitizesJobName(t *testing.T) {
	s := newTempStore(t)
	rec := store.RunRecord{
		JobName:   "my job/with:special*chars",
		Command:   "echo hi",
		StartedAt: time.Now().UTC(),
		ExitCode:  0,
		Stdout:    "hi\n",
	}
	if err := s.Save(rec); err != nil {
		t.Fatalf("Save with special chars: %v", err)
	}
}
