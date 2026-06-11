# now

Returns the current time with various accessible formats.

## Signature

`now(*, tz: str = "local") -> error|{ "date": str, "year": int, "month": int, "day": int, "weekday": int, "hour": int, "minute": int, "second": int, "time": str, "epoch": { "seconds": int, "millis": int, "nanos": int } }`

## Examples

```rad
time = now()
print("Current date:", time.date)          // -> "2024-04-05"
print("Current time:", time.time)          // -> "14:30:25"
print("Year:", time.year)                  // -> 2024

// Use epoch for timestamps
timestamp = now().epoch.seconds
print("Timestamp:", timestamp)             // -> 1712345678

// Different timezone
utc_time = now(tz="UTC")
print("UTC time:", utc_time.time)          // -> Time in UTC
```

## Category

time

## Notes

**Parameters:**

| Parameter | Type            | Description                               |
|-----------|-----------------|-------------------------------------------|
| `tz`      | `str = "local"` | Timezone (e.g., "UTC", "America/Chicago") |

Map values:

| Accessor         | Description                           | Type   | Example             |
|------------------|---------------------------------------|--------|---------------------|
| `.date`          | Current date YYYY-MM-DD               | string | 2019-12-13          |
| `.year`          | Current calendar year                 | int    | 2019                |
| `.month`         | Current calendar month                | int    | 12                  |
| `.day`           | Current calendar day                  | int    | 13                  |
| `.weekday`       | ISO day of week (Monday=1..Sunday=7)  | int    | 5                   |
| `.hour`          | Current clock hour (24h)              | int    | 14                  |
| `.minute`        | Current minute of the hour            | int    | 15                  |
| `.second`        | Current second of the minute          | int    | 16                  |
| `.time`          | Current time in "hh:mm:ss" format     | string | 14:15:16            |
| `.epoch.seconds` | Seconds since 1970-01-01 00:00:00 UTC | int    | 1576246516          |
| `.epoch.millis`  | Millis since 1970-01-01 00:00:00 UTC  | int    | 1576246516123       |
| `.epoch.nanos`   | Nanos since 1970-01-01 00:00:00 UTC   | int    | 1576246516123456789 |
