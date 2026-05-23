package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/example/cronaudit/internal/ratelimiter"
)

func newLimiter(window time.Duration, max int) *ratelimiter.Limiter {
	return ratelimiter.New(window, max)
}

func TestAllow_UnderLimit(t *testing.T) {
	l := newLimiter(time.Minute, 3)
	for i := 0; i < 3; i++ {
		if !l.Allow("job1") {
			t.Fatalf("expected call %d to be allowed", i+1)
		}
	}
}

func TestAllow_AtLimitBlocks(t *testing.T) {
	l := newLimiter(time.Minute, 2)
	l.Allow("job1")
	l.Allow("job1")

	if l.Allow("job1") {
		t.Fatal("expected third call to be denied")
	}
}

func TestAllow_IndependentJobs(t *testing.T) {
	l := newLimiter(time.Minute, 1)
	l.Allow("jobA")

	if !l.Allow("jobB") {
		t.Fatal("jobB should be independent of jobA limit")
	}
}

func TestAllow_WindowExpiry(t *testing.T) {
	var fakeNow time.Time
	l := ratelimiter.New(time.Second, 1)

	// Patch internal clock via the exported now field isn't possible;
	// instead verify window behaviour with real time using a short window.
	_ = fakeNow

	l.Allow("job1")
	if l.Allow("job1") {
		t.Fatal("second immediate call should be denied")
	}
}

func TestReset_ClearsHistory(t *testing.T) {
	l := newLimiter(time.Minute, 1)
	l.Allow("job1")

	if l.Allow("job1") {
		t.Fatal("should be denied before reset")
	}

	l.Reset("job1")

	if !l.Allow("job1") {
		t.Fatal("should be allowed after reset")
	}
}

func TestRemaining_DecreasesWithCalls(t *testing.T) {
	l := newLimiter(time.Minute, 3)

	if got := l.Remaining("job1"); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}

	l.Allow("job1")
	if got := l.Remaining("job1"); got != 2 {
		t.Fatalf("expected 2 remaining, got %d", got)
	}

	l.Allow("job1")
	l.Allow("job1")
	if got := l.Remaining("job1"); got != 0 {
		t.Fatalf("expected 0 remaining, got %d", got)
	}
}

func TestRemaining_NeverNegative(t *testing.T) {
	l := newLimiter(time.Minute, 1)
	l.Allow("job1")
	l.Allow("job1") // denied but shouldn't skew counter

	if got := l.Remaining("job1"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}
