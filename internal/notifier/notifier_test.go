package notifier_test

import (
	"strings"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/differ"
	"github.com/example/cronaudit/internal/notifier"
)

func fixedTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	return t
}

func TestNotify_InfoEvent(t *testing.T) {
	var buf strings.Builder
	n := notifier.New(&buf)
	err := n.Notify(notifier.Event{
		JobName:   "test-job",
		Level:     notifier.LevelInfo,
		Message:   "all good",
		Timestamp: fixedTime(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected INFO in output, got: %s", out)
	}
	if !strings.Contains(out, "test-job") {
		t.Errorf("expected job name in output, got: %s", out)
	}
	if !strings.Contains(out, "all good") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestNotify_DriftEvent_ShowsDiff(t *testing.T) {
	var buf strings.Builder
	n := notifier.New(&buf)
	changes := []differ.Change{
		{Op: differ.OpRemove, Line: "old line"},
		{Op: differ.OpAdd, Line: "new line"},
	}
	err := n.Notify(notifier.Event{
		JobName:   "sync-job",
		Level:     notifier.LevelDrift,
		Diff:      changes,
		Timestamp: fixedTime(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "- old line") {
		t.Errorf("expected removed line marker, got: %s", out)
	}
	if !strings.Contains(out, "+ new line") {
		t.Errorf("expected added line marker, got: %s", out)
	}
}

func TestNotifyDrift_Convenience(t *testing.T) {
	var buf strings.Builder
	n := notifier.New(&buf)
	changes := []differ.Change{
		{Op: differ.OpAdd, Line: "extra output"},
	}
	if err := n.NotifyDrift("cron-job", changes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRIFT") {
		t.Errorf("expected DRIFT level, got: %s", out)
	}
	if !strings.Contains(out, "1 line(s) changed") {
		t.Errorf("expected change count message, got: %s", out)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	// Ensure New(nil) does not panic
	n := notifier.New(nil)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}
