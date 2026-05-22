// Package pruner removes old run results from the store to prevent unbounded disk growth.
package pruner

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Pruner deletes run result files older than a configured retention period.
type Pruner struct {
	baseDir   string
	retainFor time.Duration
	logger    *log.Logger
}

// New creates a Pruner that removes files older than retainFor from baseDir.
func New(baseDir string, retainFor time.Duration, w io.Writer) *Pruner {
	if w == nil {
		w = os.Stdout
	}
	return &Pruner{
		baseDir:   baseDir,
		retainFor: retainFor,
		logger:    log.New(w, "[pruner] ", 0),
	}
}

// PruneJob removes result files for a single job that are older than the retention window.
// Files are expected to live under baseDir/<jobName>/.
func (p *Pruner) PruneJob(jobName string) (int, error) {
	dir := filepath.Join(p.baseDir, jobName)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("pruner: read dir %s: %w", dir, err)
	}

	cutoff := time.Now().Add(-p.retainFor)
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // oldest first by filename (timestamp-prefixed)

	removed := 0
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(f); err != nil {
				p.logger.Printf("failed to remove %s: %v", f, err)
				continue
			}
			removed++
		}
	}
	if removed > 0 {
		p.logger.Printf("pruned %d file(s) for job %q", removed, jobName)
	}
	return removed, nil
}

// PruneAll walks every job subdirectory under baseDir and prunes each one.
func (p *Pruner) PruneAll() (int, error) {
	entries, err := os.ReadDir(p.baseDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("pruner: read base dir %s: %w", p.baseDir, err)
	}
	total := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n, err := p.PruneJob(e.Name())
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
