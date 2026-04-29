# cronwrap

Lightweight cron job wrapper that captures stdout/stderr, reports failures to Slack or email, and tracks run history.

---

## Installation

```bash
go install github.com/yourname/cronwrap@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwrap.git && cd cronwrap && go build -o cronwrap .
```

---

## Usage

Wrap any cron command with `cronwrap` to get automatic failure reporting and history tracking.

```bash
cronwrap [options] -- <command>
```

**Example crontab entry:**

```
0 2 * * * cronwrap --slack-webhook https://hooks.slack.com/... --label "nightly-backup" -- /usr/local/bin/backup.sh
```

**Available flags:**

| Flag | Description |
|------|-------------|
| `--label` | Human-readable name for the job |
| `--slack-webhook` | Slack incoming webhook URL for failure alerts |
| `--email` | Email address to notify on failure |
| `--history-file` | Path to store run history (default: `~/.cronwrap/history.json`) |
| `--timeout` | Maximum allowed run time (e.g. `30m`) |
| `--on-success` | Also notify on successful runs |

**View run history:**

```bash
cronwrap history --label "nightly-backup"
```

---

## How It Works

1. Executes the wrapped command and captures both `stdout` and `stderr`.
2. On non-zero exit or timeout, sends an alert via Slack and/or email with the captured output.
3. Appends a record (exit code, duration, timestamp) to the local history file.

---

## License

MIT © [yourname](https://github.com/yourname)