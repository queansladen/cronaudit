// Package healthcheck exposes a simple HTTP endpoint that reports the
// liveness of the cronaudit daemon and the last-run time of each job.
package healthcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// JobStatus holds the most recent execution metadata for a single job.
type JobStatus struct {
	LastRun   time.Time `json:"last_run"`
	LastExit  int       `json:"last_exit"`
	Drifted   bool      `json:"drifted"`
}

// Server is a lightweight HTTP health-check server.
type Server struct {
	mu      sync.RWMutex
	status  map[string]JobStatus
	addr    string
	server  *http.Server
}

// New creates a new Server that will listen on addr (e.g. ":8080").
func New(addr string) *Server {
	s := &Server{
		addr:   addr,
		status: make(map[string]JobStatus),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	s.server = &http.Server{Addr: addr, Handler: mux}
	return s
}

// Record updates the stored status for a job.
func (s *Server) Record(name string, js JobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[name] = js
}

// Start begins serving in a background goroutine.
func (s *Server) Start() error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		return fmt.Errorf("healthcheck: %w", err)
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop() error {
	return s.server.Close()
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.status)
}
