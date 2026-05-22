package summarizer_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/summarizer"
)

func makeResult(name string, exit int, drifted bool, dur time.Duration, runAt time.Time) summarizer.Result {
	return summarizer.Result{
		JobName:  name,
		ExitCode: exit,
		Drifted:  drifted,
		Duration: dur,
		RunAt:    runAt,
	}
}

func TestSummarize_EmptyResults(t *testing.T) {
	s := summarizer.New(nil)
	summaries := s.Summarize(nil)
	if len(summaries) != 0 {
		t.Fatalf("expected empty summaries, got %d", len(summaries))
	}
}

func TestSummarize_SingleJob(t *testing.T) {
	now := time.Now()
	results := []summarizer.Result{
		makeResult("backup", 0, false, 200*time.Millisecond, now.Add(-2*time.Hour)),
		makeResult("backup", 1, true, 400*time.Millisecond, now.Add(-1*time.Hour)),
		makeResult("backup", 0, true, 300*time.Millisecond, now),
	}

	s := summarizer.New(nil)
	summaries := s.Summarize(results)

	js, ok := summaries["backup"]
	if !ok {
		t.Fatal("expected summary for 'backup'")
	}
	if js.TotalRuns != 3 {
		t.Errorf("TotalRuns: want 3, got %d", js.TotalRuns)
	}
	if js.FailedRuns != 1 {
		t.Errorf("FailedRuns: want 1, got %d", js.FailedRuns)
	}
	if js.DriftCount != 2 {
		t.Errorf("DriftCount: want 2, got %d", js.DriftCount)
	}
	wantAvg := 300 * time.Millisecond
	if js.AvgDuration != wantAvg {
		t.Errorf("AvgDuration: want %s, got %s", wantAvg, js.AvgDuration)
	}
	if !js.LastRun.Equal(now) {
		t.Errorf("LastRun: want %s, got %s", now, js.LastRun)
	}
}

func TestSummarize_MultipleJobs(t *testing.T) {
	now := time.Now()
	results := []summarizer.Result{
		makeResult("job-a", 0, false, 100*time.Millisecond, now),
		makeResult("job-b", 0, false, 200*time.Millisecond, now),
		makeResult("job-a", 1, true, 300*time.Millisecond, now),
	}
	s := summarizer.New(nil)
	summaries := s.Summarize(results)

	if len(summaries) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(summaries))
	}
	if summaries["job-a"].TotalRuns != 2 {
		t.Errorf("job-a TotalRuns: want 2, got %d", summaries["job-a"].TotalRuns)
	}
	if summaries["job-b"].TotalRuns != 1 {
		t.Errorf("job-b TotalRuns: want 1, got %d", summaries["job-b"].TotalRuns)
	}
}

func TestPrint_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	s := summarizer.New(&buf)
	now := time.Now()
	results := []summarizer.Result{
		makeResult("nightly", 0, false, 500*time.Millisecond, now),
	}
	summaries := s.Summarize(results)
	s.Print(summaries)

	out := buf.String()
	for _, want := range []string{"JOB", "RUNS", "FAILED", "DRIFTS", "AVG_DUR", "nightly"} {
		if !strings.Contains(out, want) {
			t.Errorf("Print output missing %q", want)
		}
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	s := summarizer.New(nil)
	if s == nil {
		t.Fatal("expected non-nil Summarizer")
	}
}
