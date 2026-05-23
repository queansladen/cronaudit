package jobcooldown_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/jobcooldown"
)

func newCooldown(gap time.Duration) *jobcooldown.Cooldown {
	return jobcooldown.New(gap)
}

func TestAllow_FirstCallAlwaysPermitted(t *testing.T) {
	c := newCooldown(10 * time.Second)
	if err := c.Allow("myjob"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_BlockedWithinCooldown(t *testing.T) {
	c := newCooldown(10 * time.Second)
	c.Record("myjob")

	if err := c.Allow("myjob"); err == nil {
		t.Fatal("expected cooldown error, got nil")
	}
}

func TestAllow_PermittedAfterCooldownExpires(t *testing.T) {
	c := newCooldown(1 * time.Millisecond)
	c.Record("myjob")
	time.Sleep(5 * time.Millisecond)

	if err := c.Allow("myjob"); err != nil {
		t.Fatalf("expected nil after cooldown, got %v", err)
	}
}

func TestAllow_IndependentJobsDoNotInterfere(t *testing.T) {
	c := newCooldown(10 * time.Second)
	c.Record("job-a")

	if err := c.Allow("job-b"); err != nil {
		t.Fatalf("job-b should not be affected by job-a cooldown: %v", err)
	}
}

func TestReset_ClearsState(t *testing.T) {
	c := newCooldown(10 * time.Second)
	c.Record("myjob")
	c.Reset("myjob")

	if err := c.Allow("myjob"); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestLastRun_ReturnsFalseWhenUnset(t *testing.T) {
	c := newCooldown(5 * time.Second)
	_, ok := c.LastRun("unknown")
	if ok {
		t.Fatal("expected false for unseen job")
	}
}

func TestLastRun_ReturnsTrueAfterRecord(t *testing.T) {
	c := newCooldown(5 * time.Second)
	before := time.Now()
	c.Record("myjob")
	after := time.Now()

	t2, ok := c.LastRun("myjob")
	if !ok {
		t.Fatal("expected true after Record")
	}
	if t2.Before(before) || t2.After(after) {
		t.Errorf("LastRun time %v not in expected range [%v, %v]", t2, before, after)
	}
}

func TestAllow_ErrorMessageContainsJobName(t *testing.T) {
	c := newCooldown(30 * time.Second)
	c.Record("special-job")

	err := c.Allow("special-job")
	if err == nil {
		t.Fatal("expected error")
	}
	if msg := err.Error(); len(msg) == 0 {
		t.Fatal("error message should not be empty")
	}
}
