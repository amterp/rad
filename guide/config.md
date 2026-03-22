---
title: Configuration
---

Rad stores its configuration in a single file: `~/.rad/config.toml`. This file is optional - if it doesn't exist, Rad uses sensible defaults for everything.

The file uses [TOML](https://toml.io) format.

## Invocation Logging

By default, Rad logs basic metadata every time you run a script. These logs power `rad check --from-logs`, which bulk-checks your recently-used scripts for issues after an upgrade.

### What Gets Logged

Each log entry records:

- **Script path** - the absolute path to the script that was run
- **Timestamp** - when the script was invoked
- **Rad version** - which version of Rad ran the script
- **Duration** - how long the script took to execute

Arguments are **not** logged by default. The log file lives at `~/.rad/logs/invocations.jsonl`.

### Using `rad check --from-logs`

After upgrading Rad, you can check whether your scripts still parse correctly:

```shell
rad check --from-logs all
```

<div class="result">
```
Checked 47 scripts: 47 passed, 0 failed (3 hints, 120 skipped).
```
</div>

This reads the invocation log to find scripts you've actually used, then checks each one. It's a quick way to catch compatibility issues before they surprise you.

The `--from-logs` flag takes a duration value that controls how far back to look. Use `all` to check everything, or a duration like `7d` to check only recently-used scripts:

```shell
rad check --from-logs 7d
```

### Settings

All settings live under `[invocation_logging]`:

```toml
[invocation_logging]
enabled = true          # Enable/disable logging (default: true)
include_args = false    # Log script arguments too (default: false)
max_size_mb = 10        # Max log file size before rotation, in MB (default: 10)
keep_rolled_logs = 2    # Rotated log files to keep (default: 2)
```

`include_args` is off by default because arguments may contain sensitive information (passwords, tokens, etc.). When rotation kicks in, Rad keeps at most `keep_rolled_logs` older copies alongside the current log file, deleting anything older.

## Summary

- Rad's config file lives at `~/.rad/config.toml` (TOML format).
- The file is optional - Rad uses defaults if it's absent.
- Invocation logging is enabled by default, powering `rad check --from-logs`.
- Only script path, timestamp, version, and duration are logged - no arguments by default.
- Log rotation is automatic, controlled by `max_size_mb` and `keep_rolled_logs`.
