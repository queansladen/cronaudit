// Package store handles persistence of cron job run results,
// enabling diff and drift detection over time.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunRecord represents a single recorded execution of a cron job.
type RunRecord struct {
	JobName   string        `json:"job_name"`
	Command   string        `json:"command"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ns"`
	ExitCode  int           `json:"exit_code"`
	Stdout    string        `json:"stdout"`
	Stderr    string        `json:"stderr"`
}

// Store persists RunRecords to a directory on disk.
type Store struct {
	dir string
}

// New creates a new Store rooted at dir, creating the directory if needed.
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("store: create directory %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a RunRecord to disk as a JSON file.
// The filename encodes the job name and timestamp for easy ordering.
func (s *Store) Save(r RunRecord) error {
	filename := fmt.Sprintf("%s_%s.json",
		sanitize(r.JobName),
		r.StartedAt.UTC().Format("20060102T150405Z"),
	)
	path := filepath.Join(s.dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("store: create file %q: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("store: encode record: %w", err)
	}
	return nil
}

// Latest returns the most recent RunRecord for the given job name,
// or nil if no records exist.
func (s *Store) Latest(jobName string) (*RunRecord, error) {
	pattern := filepath.Join(s.dir, sanitize(jobName)+"_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("store: glob %q: %w", pattern, err)
	}
	if len(matches) == 0 {
		return nil, nil
	}
	// Glob returns sorted results; last entry is the most recent.
	latest := matches[len(matches)-1]

	f, err := os.Open(latest)
	if err != nil {
		return nil, fmt.Errorf("store: open %q: %w", latest, err)
	}
	defer f.Close()

	var rec RunRecord
	if err := json.NewDecoder(f).Decode(&rec); err != nil {
		return nil, fmt.Errorf("store: decode %q: %w", latest, err)
	}
	return &rec, nil
}

// sanitize replaces characters that are unsafe in filenames with underscores.
func sanitize(s string) string {
	out := make([]byte, len(s))
	for i := range s {
		switch s[i] {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', ' ':
			out[i] = '_'
		default:
			out[i] = s[i]
		}
	}
	return string(out)
}
