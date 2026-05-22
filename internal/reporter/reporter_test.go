package reporter_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/reporter"
	"github.com/example/cronaudit/internal/store"
)

func newTempStore(t *testing.T) *store.Store {
	t.Helper()
	dir, err := os.MkdirTemp("", "cronaudit-reporter-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

func TestGenerate_EmptyStore(t *testing.T) {
	s := newTempStore(t)
	r := reporter.New(s, nil)
	now := time.Now()
	rpt, err := r.Generate([]string{"backup"}, now.Add(-time.Hour), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rpt.Summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(rpt.Summaries))
	}
	if rpt.Summaries[0].Runs != 0 {
		t.Errorf("expected 0 runs, got %d", rpt.Summaries[0].Runs)
	}
}

func TestGenerate_CountsDrifts(t *testing.T) {
	s := newTempStore(t)
	now := time.Now()

	_ = s.Save("backup", &store.Result{RunAt: now.Add(-30 * time.Minute), Output: "ok", Drifted: false})
	_ = s.Save("backup", &store.Result{RunAt: now.Add(-20 * time.Minute), Output: "changed", Drifted: true})
	_ = s.Save("backup", &store.Result{RunAt: now.Add(-10 * time.Minute), Output: "changed again", Drifted: true})

	r := reporter.New(s, nil)
	rpt, err := r.Generate([]string{"backup"}, now.Add(-time.Hour), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := rpt.Summaries[0]
	if sum.Runs != 3 {
		t.Errorf("expected 3 runs, got %d", sum.Runs)
	}
	if sum.Drifts != 2 {
		t.Errorf("expected 2 drifts, got %d", sum.Drifts)
	}
	if sum.LastOutput != "changed again" {
		t.Errorf("unexpected LastOutput: %q", sum.LastOutput)
	}
}

func TestPrint_ContainsJobName(t *testing.T) {
	s := newTempStore(t)
	var buf bytes.Buffer
	r := reporter.New(s, &buf)
	now := time.Now()

	_ = s.Save("healthcheck", &store.Result{RunAt: now.Add(-5 * time.Minute), Output: "alive", Drifted: false})

	rpt, _ := r.Generate([]string{"healthcheck"}, now.Add(-time.Hour), now)
	r.Print(rpt)

	out := buf.String()
	if !strings.Contains(out, "healthcheck") {
		t.Errorf("expected job name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "CronAudit Report") {
		t.Errorf("expected report header in output, got:\n%s", out)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	s := newTempStore(t)
	r := reporter.New(s, nil)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}
