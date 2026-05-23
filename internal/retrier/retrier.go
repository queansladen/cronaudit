// Package retrier provides retry logic with exponential backoff for cron job execution.
package retrier

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Config holds retry policy settings.
type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultConfig returns a sensible default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
}

// Retrier executes a function with exponential backoff retry logic.
type Retrier struct {
	cfg   Config
	sleep func(time.Duration)
}

// New creates a new Retrier with the given config.
func New(cfg Config) *Retrier {
	return &Retrier{
		cfg:   cfg,
		sleep: time.Sleep,
	}
}

// Attempt runs fn up to MaxAttempts times, backing off exponentially between failures.
// Returns the last error if all attempts fail, or nil on success.
func (r *Retrier) Attempt(ctx context.Context, jobName string, fn func() error) error {
	var lastErr error
	for i := 0; i < r.cfg.MaxAttempts; i++ {
		if ctx.Err() != nil {
			return fmt.Errorf("retrier: context cancelled for job %q: %w", jobName, ctx.Err())
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if i < r.cfg.MaxAttempts-1 {
			delay := r.delay(i)
			select {
			case <-ctx.Done():
				return fmt.Errorf("retrier: context cancelled for job %q: %w", jobName, ctx.Err())
			case <-time.After(delay):
			}
		}
	}
	return fmt.Errorf("retrier: job %q failed after %d attempts: %w", jobName, r.cfg.MaxAttempts, lastErr)
}

// delay calculates the backoff duration for attempt i.
func (r *Retrier) delay(attempt int) time.Duration {
	backoff := float64(r.cfg.BaseDelay) * math.Pow(r.cfg.Multiplier, float64(attempt))
	if backoff > float64(r.cfg.MaxDelay) {
		backoff = float64(r.cfg.MaxDelay)
	}
	return time.Duration(backoff)
}
