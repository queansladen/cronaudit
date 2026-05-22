package exporter_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/exporter"
	"github.com/yourorg/cronaudit/internal/store"
)

func makeResult(output string, exit int) *store.Result {
	return &store.Result{
		RunAt:    time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Output:   output,
		ExitCode: exit,
		Duration: 2*time.Second + 500*time.Millisecond,
	}
}

func TestExport_WritesValidJSONL(t *testing.T) {
	var buf bytes.Buffer
	e := exporter.New(&buf)

	err := e.Export("backup", makeResult("ok\n", 0), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var rec exporter.Record
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if rec.Job != "backup" {
		t.Errorf("job: got %q, want %q", rec.Job, "backup")
	}
	if rec.ExitCode != 0 {
		t.Errorf("exit_code: got %d, want 0", rec.ExitCode)
	}
	if rec.Drifted {
		t.Error("drifted should be false")
	}
	if rec.Duration != 2.5 {
		t.Errorf("duration: got %f, want 2.5", rec.Duration)
	}
}

func TestExport_DriftedFlag(t *testing.T) {
	var buf bytes.Buffer
	e := exporter.New(&buf)
	_ = e.Export("sync", makeResult("changed\n", 0), true)

	var rec exporter.Record
	_ = json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec)
	if !rec.Drifted {
		t.Error("expected drifted=true")
	}
}

func TestExport_NilResultReturnsError(t *testing.T) {
	var buf bytes.Buffer
	e := exporter.New(&buf)
	if err := e.Export("job", nil, false); err == nil {
		t.Error("expected error for nil result")
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	e := exporter.New(nil)
	if e == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestExportToFile_AppendsRecords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	for i := 0; i < 3; i++ {
		if err := exporter.ExportToFile(path, "cron", makeResult("out", 0), false); err != nil {
			t.Fatalf("ExportToFile error: %v", err)
		}
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}
