# cronaudit

Lightweight daemon that logs and diffs cron job output over time for drift detection.

## Installation

```bash
go install github.com/yourname/cronaudit@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronaudit.git && cd cronaudit && go build ./...
```

## Usage

Start the daemon with a config file:

```bash
cronaudit --config /etc/cronaudit/config.yaml
```

Example `config.yaml`:

```yaml
jobs:
  - name: backup
    schedule: "0 2 * * *"
    command: "/usr/local/bin/backup.sh"
  - name: cleanup
    schedule: "*/15 * * * *"
    command: "/usr/local/bin/cleanup.sh"

log_dir: /var/log/cronaudit
diff_threshold: 10
```

cronaudit will run each job on its schedule, store the output, and emit a diff when the output deviates from the previous run beyond the configured threshold.

View the drift report for a specific job:

```bash
cronaudit report --job backup --since 7d
```

## How It Works

- Runs cron jobs as a managed daemon process
- Stores output snapshots in the configured log directory
- Computes line-level diffs between consecutive runs
- Alerts (stdout, syslog, or webhook) when drift exceeds the threshold

## License

MIT