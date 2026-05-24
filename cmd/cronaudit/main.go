// Command cronaudit is a lightweight daemon that logs and diffs cron job
// output over time, enabling drift detection across scheduled task runs.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourorg/cronaudit/internal/config"
	"github.com/yourorg/cronaudit/internal/notifier"
	"github.com/yourorg/cronaudit/internal/runner"
	"github.com/yourorg/cronaudit/internal/scheduler"
	"github.com/yourorg/cronaudit/internal/store"
)

const defaultConfigPath = "cronaudit.yaml"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cronaudit: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := flag.String("config", defaultConfigPath, "path to configuration file")
	version := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *version {
		fmt.Println("cronaudit dev")
		return nil
	}

	// Load and validate configuration.
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config %q: %w", *cfgPath, err)
	}

	// Initialise the result store.
	st, err := store.New(cfg.StorePath)
	if err != nil {
		return fmt.Errorf("initialising store at %q: %w", cfg.StorePath, err)
	}
	defer st.Close()

	// Build the command runner.
	r := runner.New()

	// Build the notifier (writes to stdout by default).
	n := notifier.New(os.Stdout)

	// Assemble and start the scheduler.
	sched := scheduler.New(cfg, r, st, n)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("starting scheduler: %w", err)
	}

	log.Printf("cronaudit started — monitoring %d job(s)", len(cfg.Jobs))

	// Block until the process receives a termination signal.
	sig := waitForSignal()

	log.Printf("received signal %s — shutting down", sig)
	cancel()
	sched.Wait()

	return nil
}

// waitForSignal blocks until SIGINT or SIGTERM is received and returns the
// signal that triggered the shutdown.
func waitForSignal() os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	return <-sigCh
}
