package jobthrottle_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/example/cronaudit/internal/jobthrottle"
)

func newThrottle(max int) *jobthrottle.Throttle {
	return jobthrottle.New(max)
}

func TestAcquire_SucceedsWhenUnderLimit(t *testing.T) {
	th := newThrottle(2)
	if err := th.Acquire("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := th.InFlight("backup"); got != 1 {
		t.Fatalf("expected 1 in-flight, got %d", got)
	}
}

func TestAcquire_BlocksAtLimit(t *testing.T) {
	th := newThrottle(1)
	if err := th.Acquire("sync"); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	err := th.Acquire("sync")
	if err == nil {
		t.Fatal("expected throttle error, got nil")
	}
	if !errors.Is(err, jobthrottle.ErrThrottled) {
		t.Fatalf("expected ErrThrottled, got: %v", err)
	}
}

func TestRelease_DecrementsCount(t *testing.T) {
	th := newThrottle(1)
	_ = th.Acquire("cleanup")
	th.Release("cleanup")
	if got := th.InFlight("cleanup"); got != 0 {
		t.Fatalf("expected 0 in-flight after release, got %d", got)
	}
	// should be acquirable again
	if err := th.Acquire("cleanup"); err != nil {
		t.Fatalf("acquire after release failed: %v", err)
	}
}

func TestRelease_SafeWhenAlreadyZero(t *testing.T) {
	th := newThrottle(2)
	// releasing without acquiring should not panic or go negative
	th.Release("ghost")
	if got := th.InFlight("ghost"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAcquire_IndependentJobs(t *testing.T) {
	th := newThrottle(1)
	if err := th.Acquire("jobA"); err != nil {
		t.Fatalf("jobA acquire failed: %v", err)
	}
	// jobB has its own slot and should not be affected by jobA
	if err := th.Acquire("jobB"); err != nil {
		t.Fatalf("jobB acquire failed: %v", err)
	}
}

func TestConcurrentAcquireRelease(t *testing.T) {
	th := newThrottle(5)
	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if err := th.Acquire("concurrent"); err == nil {
				th.Release("concurrent")
			}
		}()
	}
	wg.Wait()
	if got := th.InFlight("concurrent"); got != 0 {
		t.Fatalf("expected 0 in-flight after all goroutines finish, got %d", got)
	}
}
