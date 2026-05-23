// Package deduplicator suppresses repeated identical outputs from cron jobs,
// preventing noise in drift detection when a job produces the same output
// across consecutive runs.
package deduplicator

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

// Deduplicator tracks the last seen output hash per job and reports
// whether a new result is distinct from the previous one.
type Deduplicator struct {
	mu   sync.Mutex
	hashes map[string]string
}

// New returns a new Deduplicator.
func New() *Deduplicator {
	return &Deduplicator{
		hashes: make(map[string]string),
	}
}

// IsDuplicate returns true if the given output is identical to the last
// recorded output for jobName. It always updates the stored hash.
func (d *Deduplicator) IsDuplicate(jobName, output string) bool {
	h := hash(output)

	d.mu.Lock()
	defer d.mu.Unlock()

	prev, exists := d.hashes[jobName]
	d.hashes[jobName] = h

	return exists && prev == h
}

// Reset clears the stored hash for a specific job, forcing the next
// call to IsDuplicate to return false regardless of output.
func (d *Deduplicator) Reset(jobName string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.hashes, jobName)
}

// ResetAll clears all stored hashes.
func (d *Deduplicator) ResetAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.hashes = make(map[string]string)
}

// Len returns the number of jobs currently tracked.
func (d *Deduplicator) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.hashes)
}

func hash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}
