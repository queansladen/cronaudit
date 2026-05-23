package jobqueue_test

import (
	"testing"

	"github.com/example/cronaudit/internal/jobqueue"
)

func newQueue(cap int) *jobqueue.Queue {
	return jobqueue.New(cap)
}

func TestEnqueue_AddsItem(t *testing.T) {
	q := newQueue(4)
	if err := q.Enqueue(jobqueue.Item{JobName: "backup", Priority: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestEnqueue_ReturnsErrWhenFull(t *testing.T) {
	q := newQueue(2)
	q.Enqueue(jobqueue.Item{JobName: "a", Priority: 1}) //nolint
	q.Enqueue(jobqueue.Item{JobName: "b", Priority: 1}) //nolint
	err := q.Enqueue(jobqueue.Item{JobName: "c", Priority: 1})
	if err != jobqueue.ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
}

func TestDequeue_ReturnsErrWhenEmpty(t *testing.T) {
	q := newQueue(4)
	_, err := q.Dequeue()
	if err != jobqueue.ErrQueueEmpty {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestDequeue_HighestPriorityFirst(t *testing.T) {
	q := newQueue(8)
	q.Enqueue(jobqueue.Item{JobName: "low", Priority: 1})    //nolint
	q.Enqueue(jobqueue.Item{JobName: "high", Priority: 10})  //nolint
	q.Enqueue(jobqueue.Item{JobName: "medium", Priority: 5}) //nolint

	expected := []string{"high", "medium", "low"}
	for _, name := range expected {
		item, err := q.Dequeue()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.JobName != name {
			t.Errorf("expected %q, got %q", name, item.JobName)
		}
	}
}

func TestDequeue_FIFOForEqualPriority(t *testing.T) {
	q := newQueue(8)
	q.Enqueue(jobqueue.Item{JobName: "first", Priority: 5})  //nolint
	q.Enqueue(jobqueue.Item{JobName: "second", Priority: 5}) //nolint
	q.Enqueue(jobqueue.Item{JobName: "third", Priority: 5})  //nolint

	for _, want := range []string{"first", "second", "third"} {
		got, _ := q.Dequeue()
		if got.JobName != want {
			t.Errorf("expected %q, got %q", want, got.JobName)
		}
	}
}

func TestDrain_ReturnsAllAndClearsQueue(t *testing.T) {
	q := newQueue(8)
	q.Enqueue(jobqueue.Item{JobName: "a", Priority: 3}) //nolint
	q.Enqueue(jobqueue.Item{JobName: "b", Priority: 1}) //nolint
	q.Enqueue(jobqueue.Item{JobName: "c", Priority: 2}) //nolint

	items := q.Drain()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].JobName != "a" {
		t.Errorf("expected first item to be 'a', got %q", items[0].JobName)
	}
	if q.Len() != 0 {
		t.Errorf("expected queue to be empty after drain, got len %d", q.Len())
	}
}

func TestNew_DefaultCapacityWhenZero(t *testing.T) {
	q := jobqueue.New(0)
	for i := 0; i < 64; i++ {
		if err := q.Enqueue(jobqueue.Item{JobName: "x", Priority: 0}); err != nil {
			t.Fatalf("unexpected error at item %d: %v", i, err)
		}
	}
	if err := q.Enqueue(jobqueue.Item{JobName: "overflow", Priority: 0}); err != jobqueue.ErrQueueFull {
		t.Errorf("expected ErrQueueFull at default capacity, got %v", err)
	}
}
