package joblock_test

import (
	"sync"
	"testing"

	"github.com/example/cronaudit/internal/joblock"
)

func newLocker() *joblock.Locker {
	return joblock.New()
}

func TestAcquire_SucceedsWhenFree(t *testing.T) {
	l := newLocker()
	release, err := l.Acquire("myjob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if release == nil {
		t.Fatal("expected a release func, got nil")
	}
	release()
}

func TestAcquire_FailsWhenAlreadyLocked(t *testing.T) {
	l := newLocker()
	release, err := l.Acquire("myjob")
	if err != nil {
		t.Fatalf("unexpected error on first acquire: %v", err)
	}
	defer release()

	_, err2 := l.Acquire("myjob")
	if err2 == nil {
		t.Fatal("expected error on second acquire, got nil")
	}
}

func TestAcquire_IndependentJobsDoNotBlock(t *testing.T) {
	l := newLocker()
	rel1, err1 := l.Acquire("job-a")
	if err1 != nil {
		t.Fatalf("job-a: %v", err1)
	}
	defer rel1()

	rel2, err2 := l.Acquire("job-b")
	if err2 != nil {
		t.Fatalf("job-b: %v", err2)
	}
	defer rel2()
}

func TestIsLocked_ReflectsState(t *testing.T) {
	l := newLocker()
	if l.IsLocked("myjob") {
		t.Fatal("expected unlocked before acquire")
	}

	release, _ := l.Acquire("myjob")
	if !l.IsLocked("myjob") {
		t.Fatal("expected locked after acquire")
	}

	release()
	if l.IsLocked("myjob") {
		t.Fatal("expected unlocked after release")
	}
}

func TestActiveJobs_ReturnsHeldLocks(t *testing.T) {
	l := newLocker()
	rel1, _ := l.Acquire("alpha")
	rel2, _ := l.Acquire("beta")
	defer rel1()
	defer rel2()

	active := l.ActiveJobs()
	if len(active) != 2 {
		t.Fatalf("expected 2 active jobs, got %d", len(active))
	}
}

func TestAcquire_ConcurrentSafeRelease(t *testing.T) {
	l := newLocker()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := l.Acquire("shared")
			if err == nil {
				release()
			}
		}()
	}
	wg.Wait()

	if l.IsLocked("shared") {
		t.Fatal("lock should be free after all goroutines finish")
	}
}
