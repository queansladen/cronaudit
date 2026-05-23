package deduplicator_test

import (
	"testing"

	"github.com/yourorg/cronaudit/internal/deduplicator"
)

func newDedup() *deduplicator.Deduplicator {
	return deduplicator.New()
}

func TestIsDuplicate_FirstCallIsNeverDuplicate(t *testing.T) {
	d := newDedup()
	if d.IsDuplicate("job1", "some output") {
		t.Fatal("first call should never be a duplicate")
	}
}

func TestIsDuplicate_SameOutputIsDuplicate(t *testing.T) {
	d := newDedup()
	d.IsDuplicate("job1", "hello")
	if !d.IsDuplicate("job1", "hello") {
		t.Fatal("identical output should be detected as duplicate")
	}
}

func TestIsDuplicate_DifferentOutputIsNotDuplicate(t *testing.T) {
	d := newDedup()
	d.IsDuplicate("job1", "hello")
	if d.IsDuplicate("job1", "world") {
		t.Fatal("different output should not be a duplicate")
	}
}

func TestIsDuplicate_IndependentJobs(t *testing.T) {
	d := newDedup()
	d.IsDuplicate("job1", "same")
	d.IsDuplicate("job2", "same")

	// Both jobs have the same output but are tracked independently.
	if !d.IsDuplicate("job1", "same") {
		t.Fatal("job1 should report duplicate")
	}
	if !d.IsDuplicate("job2", "same") {
		t.Fatal("job2 should report duplicate")
	}
}

func TestReset_ForcesNonDuplicate(t *testing.T) {
	d := newDedup()
	d.IsDuplicate("job1", "hello")
	d.Reset("job1")
	if d.IsDuplicate("job1", "hello") {
		t.Fatal("after reset, same output should not be a duplicate")
	}
}

func TestResetAll_ClearsAllJobs(t *testing.T) {
	d := newDedup()
	d.IsDuplicate("job1", "a")
	d.IsDuplicate("job2", "b")
	if d.Len() != 2 {
		t.Fatalf("expected 2 tracked jobs, got %d", d.Len())
	}

	d.ResetAll()
	if d.Len() != 0 {
		t.Fatalf("expected 0 tracked jobs after ResetAll, got %d", d.Len())
	}
}

func TestLen_TracksCorrectly(t *testing.T) {
	d := newDedup()
	if d.Len() != 0 {
		t.Fatal("expected empty deduplicator")
	}
	d.IsDuplicate("job1", "x")
	d.IsDuplicate("job1", "y") // same job, should not increase count
	d.IsDuplicate("job2", "x")
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
}
