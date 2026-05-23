// Package healthcheck provides a lightweight HTTP server embedded in the
// cronaudit daemon. It exposes two endpoints:
//
//   - GET /healthz  — returns 200 OK when the daemon is alive.
//   - GET /status   — returns a JSON map of each monitored job's last-run
//     time, exit code, and whether drift was detected.
//
// The server is optional; it is only started when the "healthcheck.addr"
// field is present in the cronaudit configuration file.
package healthcheck
