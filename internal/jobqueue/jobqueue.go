// Package jobqueue provides a bounded, priority-aware queue for scheduling
// cron job executions. Jobs with higher priority are dequeued first; equal
// priority jobs are ordered FIFO.
package jobqueue

import (
	"errors"
	"sync"
)

// ErrQueueFull is returned when an Enqueue call would exceed the queue capacity.
var ErrQueueFull = errors.New("jobqueue: queue is full")

// ErrQueueEmpty is returned when Dequeue is called on an empty queue.
var ErrQueueEmpty = errors.New("jobqueue: queue is empty")

// Item represents a single queued job execution request.
type Item struct {
	JobName  string
	Priority int // higher value = higher priority
}

// Queue is a thread-safe bounded priority queue.
type Queue struct {
	mu       sync.Mutex
	items    []Item
	capacity int
}

// New creates a Queue with the given maximum capacity.
func New(capacity int) *Queue {
	if capacity <= 0 {
		capacity = 64
	}
	return &Queue{
		items:    make([]Item, 0, capacity),
		capacity: capacity,
	}
}

// Enqueue adds an item to the queue, maintaining priority order.
// Returns ErrQueueFull if the queue has reached its capacity.
func (q *Queue) Enqueue(item Item) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.capacity {
		return ErrQueueFull
	}

	// Insert in descending priority order (binary search position).
	pos := len(q.items)
	for i, existing := range q.items {
		if item.Priority > existing.Priority {
			pos = i
			break
		}
	}
	q.items = append(q.items, Item{})
	copy(q.items[pos+1:], q.items[pos:])
	q.items[pos] = item
	return nil
}

// Dequeue removes and returns the highest-priority item.
// Returns ErrQueueEmpty if the queue is empty.
func (q *Queue) Dequeue() (Item, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return Item{}, ErrQueueEmpty
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

// Len returns the current number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Drain removes and returns all items in priority order.
func (q *Queue) Drain() []Item {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]Item, len(q.items))
	copy(out, q.items)
	q.items = q.items[:0]
	return out
}
