// Package watcher monitors the cronaudit configuration file for changes
// and signals the daemon to reload its job schedule without restarting.
package watcher

import (
	"context"
	"log"
	"os"
	"time"
)

// ReloadFunc is called when a configuration change is detected.
type ReloadFunc func() error

// Watcher polls a file for modifications and triggers a reload callback.
type Watcher struct {
	path     string
	interval time.Duration
	onReload ReloadFunc
	logger   *log.Logger
	lastMod  time.Time
}

// New creates a Watcher that checks path every interval duration.
// When the file's modification time changes, onReload is invoked.
// If logger is nil, log.Default() is used.
func New(path string, interval time.Duration, onReload ReloadFunc, logger *log.Logger) *Watcher {
	if logger == nil {
		logger = log.Default()
	}
	return &Watcher{
		path:     path,
		interval: interval,
		onReload: onReload,
		logger:   logger,
	}
}

// Start begins polling in the background. It returns immediately and stops
// when ctx is cancelled. The initial modification time is recorded so that
// only genuine changes trigger a reload after startup.
func (w *Watcher) Start(ctx context.Context) {
	info, err := os.Stat(w.path)
	if err != nil {
		w.logger.Printf("watcher: could not stat %q on start: %v", w.path, err)
	} else {
		w.lastMod = info.ModTime()
	}

	go w.loop(ctx)
}

// loop is the internal polling goroutine.
func (w *Watcher) loop(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Printf("watcher: stopping poll for %q", w.path)
			return
		case <-ticker.C:
			w.check()
		}
	}
}

// check stats the watched file and calls onReload if it has been modified.
func (w *Watcher) check() {
	info, err := os.Stat(w.path)
	if err != nil {
		// File may have been temporarily removed during an atomic write.
		w.logger.Printf("watcher: stat %q: %v", w.path, err)
		return
	}

	mod := info.ModTime()
	if !mod.After(w.lastMod) {
		return
	}

	w.logger.Printf("watcher: detected change in %q (mtime %s → %s), reloading",
		w.path, w.lastMod.Format(time.RFC3339), mod.Format(time.RFC3339))

	w.lastMod = mod

	if err := w.onReload(); err != nil {
		w.logger.Printf("watcher: reload error: %v", err)
	}
}
