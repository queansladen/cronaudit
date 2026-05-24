package jobauditlog_test

import (
	"sync"
	"testing"

	"github.com/cronaudit/cronaudit/internal/jobauditlog"
)

// TestConcurrentRecord verifies that concurrent writes from multiple goroutines
// do not cause data races and that all entries are eventually recorded.
func TestConcurrentRecord(t *testing.T) {
	const workers = 20
	const perWorker = 50

	l := jobauditlog.New(0)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				l.Record("job", jobauditlog.EventSuccess, "")
			}
		}(i)
	}
	wg.Wait()

	if l.Len() != workers*perWorker {
		t.Fatalf("expected %d entries, got %d", workers*perWorker, l.Len())
	}
}

// TestBoundedLog_NeverExceedsMax records many events and asserts the cap holds.
func TestBoundedLog_NeverExceedsMax(t *testing.T) {
	const max = 10
	l := jobauditlog.New(max)

	for i := 0; i < 100; i++ {
		l.Record("job", jobauditlog.EventDrift, "diff")
		if l.Len() > max {
			t.Fatalf("log exceeded max size %d (got %d) after %d inserts", max, l.Len(), i+1)
		}
	}
}
