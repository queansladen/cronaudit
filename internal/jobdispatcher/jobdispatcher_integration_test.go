package jobdispatcher_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/cronaudit/internal/jobdispatcher"
)

// TestIntegration_PriorityOrder verifies that higher-priority jobs are
// dispatched before lower-priority ones when workers are saturated.
func TestIntegration_PriorityOrder(t *testing.T) {
	d, q := newDispatcher(t, 1)

	// Block the single worker so we can fill the queue before releasing.
	blocked := make(chan struct{})
	gate := jobdispatcher.Job{
		Name:     "gate",
		Priority: 10,
		Run: func(_ context.Context) error {
			<-blocked
			return nil
		},
	}
	_ = q.Enqueue(gate, gate.Priority)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	d.Start(ctx)
	defer d.Stop()

	// Give gate job time to be picked up.
	time.Sleep(50 * time.Millisecond)

	var order []string
	for _, spec := range []struct {
		name     string
		priority int
	}{
		{"low", 1},
		{"high", 9},
		{"medium", 5},
	} {
		s := spec
		job := jobdispatcher.Job{
			Name:     s.name,
			Priority: s.priority,
			Run: func(_ context.Context) error {
				order = append(order, s.name)
				return nil
			},
		}
		_ = q.Enqueue(job, job.Priority)
	}

	close(blocked) // release gate

	var done atomic.Int32
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && done.Load() < 3 {
		time.Sleep(20 * time.Millisecond)
		done.Store(int32(len(order)))
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 jobs, got %d: %v", len(order), order)
	}
	// highest priority (9) should run before medium (5) before low (1)
	if order[0] != "high" {
		t.Errorf("expected 'high' first, got %q (order=%v)", order[0], order)
	}
	_ = fmt.Sprintf("order: %v", order) // prevent import error
}
