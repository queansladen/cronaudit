package healthcheck_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/healthcheck"
)

func newTestServer(t *testing.T) (*healthcheck.Server, string) {
	t.Helper()
	srv := healthcheck.New("127.0.0.1:0")
	// Use a fixed high port for tests to avoid races with :0 and http.Server.
	// In practice a real integration test would use httptest; here we start
	// the embedded server on a known free port.
	srv2 := healthcheck.New("127.0.0.1:19876")
	if err := srv2.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = srv2.Stop() })
	_ = srv
	return srv2, "http://127.0.0.1:19876"
}

func TestHealthz_Returns200(t *testing.T) {
	_, base := newTestServer(t)
	resp, err := http.Get(base + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("expected body 'ok', got %q", body)
	}
}

func TestStatus_EmptyOnStart(t *testing.T) {
	_, base := newTestServer(t)
	resp, err := http.Get(base + "/status")
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]healthcheck.JobStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}

func TestStatus_ReflectsRecordedJob(t *testing.T) {
	srv, base := newTestServer(t)
	now := time.Now().UTC().Truncate(time.Second)
	srv.Record("backup", healthcheck.JobStatus{
		LastRun:  now,
		LastExit: 0,
		Drifted:  true,
	})
	resp, err := http.Get(base + "/status")
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]healthcheck.JobStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	js, ok := result["backup"]
	if !ok {
		t.Fatal("expected 'backup' key in status")
	}
	if !js.Drifted {
		t.Error("expected Drifted=true")
	}
	if js.LastExit != 0 {
		t.Errorf("expected LastExit=0, got %d", js.LastExit)
	}
}
