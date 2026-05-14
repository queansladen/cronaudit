package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job represents a single cron job to audit.
type Job struct {
	Name     string        `yaml:"name"`
	Command  string        `yaml:"command"`
	Schedule string        `yaml:"schedule"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Config holds the full cronaudit configuration.
type Config struct {
	LogDir  string `yaml:"log_dir"`
	DiffDir string `yaml:"diff_dir"`
	Jobs    []Job  `yaml:"jobs"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		LogDir:  "/var/log/cronaudit/runs",
		DiffDir: "/var/log/cronaudit/diffs",
	}
}

// Load reads a YAML config file from path and returns a validated Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// validate checks required fields and applies defaults.
func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("no jobs defined")
	}
	for i, j := range c.Jobs {
		if j.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if j.Command == "" {
			return fmt.Errorf("job %q: command is required", j.Name)
		}
		if j.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", j.Name)
		}
		if j.Timeout == 0 {
			c.Jobs[i].Timeout = 30 * time.Second
		}
	}
	return nil
}
