package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronaudit-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
log_dir: /tmp/logs
diff_dir: /tmp/diffs
jobs:
  - name: disk-check
    command: df -h
    schedule: "*/5 * * * *"
    timeout: 10s
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogDir != "/tmp/logs" {
		t.Errorf("expected log_dir /tmp/logs, got %q", cfg.LogDir)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Jobs[0].Timeout)
	}
}

func TestLoad_DefaultTimeout(t *testing.T) {
	content := `
jobs:
  - name: uptime
    command: uptime
    schedule: "@hourly"
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jobs[0].Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Jobs[0].Timeout)
	}
}

func TestLoad_MissingCommand(t *testing.T) {
	content := `
jobs:
  - name: broken
    schedule: "@daily"
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing command, got nil")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	content := `log_dir: /tmp/logs\n`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty jobs list, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
