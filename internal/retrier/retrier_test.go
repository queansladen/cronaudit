package retrier_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronaudit/cronaudit/internal/retrier"
)

func newFastRetrier(maxAttempts int) *retrier.Retrier {
	return retrier.New(retrier.Config{
		MaxAttempts: maxAttempts,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
	})
}

func TestAttempt_SuccessOnFirstTry(t *testing.T) {
	r := newFastRetrier(3)
	calls := 0
	err := r.Attempt(context.Background(), "job1", func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestAttempt_RetriesOnFailure(t *testing.T) {
	r := newFastRetrier(3)
	calls := 0
	err := r.Attempt(context.Background(), "job2", func() error {
		calls++
		if calls < 3 {
			return errors.New("transient error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestAttempt_ExhaustsAllAttempts(t *testing.T) {
	r := newFastRetrier(3)
	calls := 0
	err := r.Attempt(context.Background(), "job3", func() error {
		calls++
		return errors.New("persistent error")
	})
	if err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestAttempt_ContextCancelledBeforeStart(t *testing.T) {
	r := newFastRetrier(3)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.Attempt(ctx, "job4", func() error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestAttempt_ContextCancelledDuringBackoff(t *testing.T) {
	r := retrier.New(retrier.Config{
		MaxAttempts: 5,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Multiplier:  2.0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := r.Attempt(ctx, "job5", func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error due to context timeout during backoff")
	}
}

func TestDefaultConfig_SaneValues(t *testing.T) {
	cfg := retrier.DefaultConfig()
	if cfg.MaxAttempts <= 0 {
		t.Errorf("MaxAttempts should be positive, got %d", cfg.MaxAttempts)
	}
	if cfg.BaseDelay <= 0 {
		t.Errorf("BaseDelay should be positive, got %v", cfg.BaseDelay)
	}
	if cfg.MaxDelay < cfg.BaseDelay {
		t.Errorf("MaxDelay should be >= BaseDelay")
	}
}
