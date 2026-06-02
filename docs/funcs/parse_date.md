# parse_date

Parses a date string into the same time map format as [`now()`](#now) and [`parse_epoch()`](#parse_epoch).

## Signature

`parse_date(_date: str, *, format: str?, tz: str = "local") -> error|{ "date": str, "year": int, "month": int, "day": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }`

## Examples

```rad
// Auto-detect common formats
time = parse_date("2026-03-22")
print(time.date)              // -> "2026-03-22"
print(time.epoch.seconds)     // -> epoch at midnight local time

time = parse_date("2026-03-22T14:30:00Z")
print(time.hour)              // -> 14 (or local equivalent)

// Custom format for non-standard date strings
time = parse_date("22/03/2026", format="DD/MM/YYYY")
print(time.date)              // -> "2026-03-22"

time = parse_date("22.03.2026 14:30", format="DD.MM.YYYY HH:mm")
print(time.hour, time.minute) // -> 14 30

// Timezone conversion
time = parse_date("2026-03-22T14:30:00Z", tz="America/Chicago")
print(time.hour)              // -> 9 (CDT = UTC-5)

// Error handling
time = parse_date("bad input") catch:
    print(time)  // error message with format hints
```

## Category

time

## Notes

**Parameters:**

| Parameter | Type            | Description                                            |
|-----------|-----------------|--------------------------------------------------------|
| `_date`   | `str`           | The date string to parse                               |
| `format`  | `str?`          | Format string using tokens (see below). Auto-detects if omitted |
| `tz`      | `str = "local"` | Timezone (e.g., "UTC", "America/Chicago")              |

**Auto-detected formats** (when `format` is omitted):

- `YYYY-MM-DD` (e.g., `2026-03-22`)
- `YYYY-MM-DDTHH:mm:ss` (e.g., `2026-03-22T14:30:00`)
- `YYYY-MM-DDTHH:mm:ssZ` or with offset (e.g., `2026-03-22T14:30:00+05:00`)
- `YYYY-MM-DD HH:mm:ss` (space-separated)
- All of the above with optional fractional seconds (e.g., `.123456`)

**Format tokens** (for the `format` parameter):

| Token  | Meaning              | Example |
|--------|----------------------|---------|
| `YYYY` | 4-digit year         | `2026`  |
| `MM`   | 2-digit month (01-12)| `03`    |
| `DD`   | 2-digit day (01-31)  | `22`    |
| `HH`   | 2-digit hour, 24h    | `14`    |
| `mm`   | 2-digit minute       | `30`    |
| `ss`   | 2-digit second       | `00`    |

All other characters in the format string are treated as literal separators. Format strings should
contain only tokens and separators - avoid embedding prose text, as tokens like `mm` and `ss` will be
matched inside words (e.g., "co**mm**it" or "acce**ss**").

Note: `MM` (uppercase) is **month**, `mm` (lowercase) is **minute**. Mixing these up will produce
wrong results or parse errors.

For strings without timezone information, the time is interpreted in the `tz` timezone (local by
default). For strings with timezone info (e.g., `Z`, `+05:00`), the time is parsed in that timezone
and then converted to the output `tz`.
