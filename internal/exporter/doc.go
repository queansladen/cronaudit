// Package exporter provides JSON Lines (JSONL) export of cronaudit run
// results, enabling downstream ingestion by log aggregators, alerting
// pipelines, or custom dashboards.
//
// Each line written by [Exporter.Export] is a self-contained JSON object
// containing the job name, timestamp, exit code, duration, drift flag, and
// captured output.
package exporter
