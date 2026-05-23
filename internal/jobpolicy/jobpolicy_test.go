package jobpolicy_test

import (
	"testing"
	"time"

	"github.com/example/cronaudit/internal/jobpolicy"
)

// monday returns a fixed Monday at 10:00 UTC.
func monday() time.Time {
	return time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC) // Monday
}

func TestAllow_NoPolicyAlwaysPermits(t *testing.T) {
	e := jobpolicy.New(nil)
	if err := e.Allow("backup", monday()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_AllowedDay(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{JobName: "backup", AllowedDays: []time.Weekday{time.Monday, time.Wednesday}},
	})
	if err := e.Allow("backup", monday()); err != nil {
		t.Fatalf("expected nil on allowed day, got %v", err)
	}
}

func TestAllow_BlockedDay(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{JobName: "backup", AllowedDays: []time.Weekday{time.Wednesday}},
	})
	if err := e.Allow("backup", monday()); err == nil {
		t.Fatal("expected error on blocked day")
	}
}

func TestAllow_WithinWindow(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{
			JobName:     "report",
			WindowStart: 9 * time.Hour,
			WindowEnd:   17 * time.Hour,
		},
	})
	// monday() is 10:00 — inside [09:00, 17:00)
	if err := e.Allow("report", monday()); err != nil {
		t.Fatalf("expected nil inside window, got %v", err)
	}
}

func TestAllow_OutsideWindow(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{
			JobName:     "report",
			WindowStart: 9 * time.Hour,
			WindowEnd:   10 * time.Hour,
		},
	})
	// monday() is exactly 10:00 — outside [09:00, 10:00)
	if err := e.Allow("report", monday()); err == nil {
		t.Fatal("expected error outside window")
	}
}

func TestAllow_BlockedAfterMaxConsecutiveFailures(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{JobName: "sync", MaxConsecutiveFailures: 3},
	})
	for i := 0; i < 3; i++ {
		_ = e.RecordResult("sync", 1)
	}
	if err := e.Allow("sync", monday()); err == nil {
		t.Fatal("expected block after 3 consecutive failures")
	}
}

func TestAllow_ResetAfterSuccess(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{JobName: "sync", MaxConsecutiveFailures: 2},
	})
	_ = e.RecordResult("sync", 1)
	_ = e.RecordResult("sync", 1)
	_ = e.RecordResult("sync", 0) // success resets counter
	if err := e.Allow("sync", monday()); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestRecordResult_UnknownJobReturnsError(t *testing.T) {
	e := jobpolicy.New(nil)
	if err := e.RecordResult("ghost", 1); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestConsecutiveFailures_TracksCount(t *testing.T) {
	e := jobpolicy.New([]jobpolicy.Policy{
		{JobName: "etl", MaxConsecutiveFailures: 5},
	})
	_ = e.RecordResult("etl", 1)
	_ = e.RecordResult("etl", 1)
	if got := e.ConsecutiveFailures("etl"); got != 2 {
		t.Fatalf("expected 2 consecutive failures, got %d", got)
	}
}
