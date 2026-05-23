package jobcooldown_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/jobcooldown"
)

// TestConcurrentAccess verifies that concurrent Record and Allow calls for
// multiple jobs do not cause data races (run with -race).
func TestConcurrentAccess(t *testing.T) {
	c := jobcooldown.New(50 * time.Millisecond)
	jobs := []string{"alpha", "beta", "gamma"}

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		for _, name := range jobs {
			wg.Add(2)
			go func(n string) {
				defer wg.Done()
				_ = c.Allow(n)
			}(name)
			go func(n string) {
				defer wg.Done()
				c.Record(n)
			}(name)
		}
	}
	wg.Wait()
}

// TestRecordThenAllowCycle simulates a realistic run-check-run cycle.
func TestRecordThenAllowCycle(t *testing.T) {
	gap := 20 * time.Millisecond
	c := jobcooldown.New(gap)

	for cycle := 0; cycle < 3; cycle++ {
		if err := c.Allow("cyclejob"); err != nil {
			t.Fatalf("cycle %d: expected allow before record, got %v", cycle, err)
		}
		c.Record("cyclejob")
		if err := c.Allow("cyclejob"); err == nil {
			t.Fatalf("cycle %d: expected block immediately after record", cycle)
		}
		time.Sleep(gap + 5*time.Millisecond)
	}
}
