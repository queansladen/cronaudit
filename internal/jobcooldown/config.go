package jobcooldown

import (
	"errors"
	"time"
)

// Config holds the parameters used to construct a Cooldown.
type Config struct {
	// MinGap is the minimum duration that must elapse between successive runs
	// of the same job. Must be greater than zero.
	MinGap time.Duration `yaml:"min_gap"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MinGap: 30 * time.Second,
	}
}

// Validate returns an error if the Config contains invalid values.
func (c Config) Validate() error {
	if c.MinGap <= 0 {
		return errors.New("jobcooldown: min_gap must be greater than zero")
	}
	return nil
}

// NewFromConfig constructs a Cooldown from a validated Config.
func NewFromConfig(cfg Config) (*Cooldown, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return New(cfg.MinGap), nil
}
