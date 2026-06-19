# parse_epoch

Parses a Unix epoch timestamp into various time formats.

```rad
parse_epoch(_epoch: int|float, *, tz: str = "local", unit: ["auto", "seconds", "millis", "micros", "nanos", "milliseconds", "microseconds", "nanoseconds"] = "auto") -> error|{ "date": str, "year": int, "month": int, "day": int, "weekday": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }
```

```rad
// Parse seconds epoch (auto-detected)
time = parse_epoch(1712345678)
print(time.date, time.time)  // -> "2024-04-05 22:01:18"

// Parse milliseconds with timezone
time = parse_epoch(1712345678123, tz="America/Chicago")
print(time.hour)  // -> Hour in Chicago timezone

// Explicit unit specification
time = parse_epoch(1712345678000, unit="millis")

// Float epoch with sub-second precision
time = parse_epoch(1712345678.5)  // 1712345678 seconds + 500ms
print(time.epoch.millis)  // -> 1712345678500

// Float with explicit unit (sub-millisecond precision)
time = parse_epoch(1712345678123.25, unit="millis")
print(time.epoch.nanos)  // -> 1712345678123250000

// Error handling
time, err = parse_epoch(1712345678, tz="Invalid/Timezone")
if err:
    print("Invalid timezone:", err.msg)
```

## Notes

**Parameters:**

| Parameter | Type                                                        | Description                                         |
| --------- | ----------------------------------------------------------- | --------------------------------------------------- |
| `_epoch`  | `int|float`                                                 | Unix epoch timestamp (float for sub-unit precision) |
| `tz`      | `str = "local"`                                             | Timezone (e.g., "UTC", "America/Chicago")           |
| `unit`    | `["auto", "seconds", "millis", "micros", "nanos"] = "auto"` | Timestamp unit (auto-detects by default)            |

Converts an epoch timestamp to the same format as `now()`. Auto-detects units from digit count, or specify
explicitly. When using a float, the fractional part provides sub-unit precision (e.g., `1712345678.5` seconds includes
500 milliseconds).
