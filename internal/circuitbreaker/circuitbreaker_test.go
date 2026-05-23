package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/example/cronaudit/internal/circuitbreaker"
)

func newBreaker(threshold int) *circuitbreaker.Breaker {
	return circuitbreaker.New(threshold, 50*time.Millisecond)
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := newBreaker(3)
	if err := b.Allow("myjob"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_OpensAfterThreshold(t *testing.T) {
	b := newBreaker(2)
	b.RecordFailure("job")
	b.RecordFailure("job")

	if err := b.Allow("job"); err == nil {
		t.Fatal("expected circuit to be open")
	}
}

func TestAllow_BelowThresholdStillClosed(t *testing.T) {
	b := newBreaker(3)
	b.RecordFailure("job")
	b.RecordFailure("job")

	if err := b.Allow("job"); err != nil {
		t.Fatalf("expected circuit closed, got %v", err)
	}
}

func TestRecordSuccess_ResetsClosed(t *testing.T) {
	b := newBreaker(2)
	b.RecordFailure("job")
	b.RecordFailure("job")
	b.RecordSuccess("job")

	if err := b.Allow("job"); err != nil {
		t.Fatalf("expected circuit closed after success, got %v", err)
	}
	if b.State("job") != circuitbreaker.StateClosed {
		t.Errorf("expected StateClosed, got %s", b.State("job"))
	}
}

func TestAllow_HalfOpenAfterRecoveryWait(t *testing.T) {
	b := circuitbreaker.New(1, 30*time.Millisecond)
	b.RecordFailure("job")

	if err := b.Allow("job"); err == nil {
		t.Fatal("expected open circuit immediately after failure")
	}

	time.Sleep(40 * time.Millisecond)

	if err := b.Allow("job"); err != nil {
		t.Fatalf("expected half-open after recovery wait, got %v", err)
	}
	if b.State("job") != circuitbreaker.StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %s", b.State("job"))
	}
}

func TestAllow_IndependentJobs(t *testing.T) {
	b := newBreaker(2)
	b.RecordFailure("jobA")
	b.RecordFailure("jobA")

	if err := b.Allow("jobB"); err != nil {
		t.Fatalf("jobB should be unaffected by jobA failures, got %v", err)
	}
}

func TestState_DefaultIsClosed(t *testing.T) {
	b := newBreaker(3)
	if s := b.State("unknown"); s != circuitbreaker.StateClosed {
		t.Errorf("expected StateClosed for unseen job, got %s", s)
	}
}

func TestAllow_ReopensAfterFailureInHalfOpen(t *testing.T) {
	// After the recovery window expires the circuit enters half-open and
	// allows a single probe. If that probe fails the circuit should return
	// to open immediately.
	b := circuitbreaker.New(1, 30*time.Millisecond)
	b.RecordFailure("job")

	time.Sleep(40 * time.Millisecond)

	// Probe allowed in half-open state.
	if err := b.Allow("job"); err != nil {
		t.Fatalf("expected half-open probe to be allowed, got %v", err)
	}

	// Probe fails — circuit should reopen.
	b.RecordFailure("job")

	if err := b.Allow("job"); err == nil {
		t.Fatal("expected circuit to reopen after half-open probe failure")
	}
	if b.State("job") != circuitbreaker.StateOpen {
		t.Errorf("expected StateOpen after probe failure, got %s", b.State("job"))
	}
}
