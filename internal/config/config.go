package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultTimeout is applied to any job that does not specify its own timeout.
const DefaultTimeout = 30 * time.Second

// Job describes a single cron job to audit.
type Job struct {
	Name    string        `yaml:"name"`
	Command string        `yaml:"command"`
	Schedule string       `yaml:"schedule"`
	Timeout time.Duration `yaml:"timeout"`
}

// Config is the top-level configuration for cronaudit.
type Config struct {
	LogDir         string        `yaml:"log_dir"`
	DefaultTimeout time.Duration `yaml:"default_timeout"`
	Jobs           []Job         `yaml:"jobs"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		LogDir:         "/var/log/cronaudit",
		DefaultTimeout: DefaultTimeout,
	}
}

// Load reads and validates a YAML config file from the given path.
// Missing optional fields are filled from DefaultConfig.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Jobs) == 0 {
		return errors.New("config must define at least one job")
	}

	seen := make(map[string]bool)
	for i, job := range cfg.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Command == "" {
			return fmt.Errorf("job %q: command is required", job.Name)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if seen[job.Name] {
			return fmt.Errorf("duplicate job name: %q", job.Name)
		}
		seen[job.Name] = true
	}

	return nil
}
