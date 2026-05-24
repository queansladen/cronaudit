// Package jobpause allows individual cron jobs to be temporarily suspended
// without altering the configuration file. A paused job is skipped by the
// scheduler until the pause window expires or Resume is called explicitly.
//
// Pauses are held in memory only; they do not survive a process restart.
package jobpause
