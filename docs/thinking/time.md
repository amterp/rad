# Time

## Parsing Time

In e.g. 1748858179

1. now(tz: string = <local>) -> map
2. parse_epoch(epoch: int, tz: string = <local>) -> map

map looks like:

```json
{
  "date": "2025-06-02",
  "day": 2,
  "epoch": {
    "millis": 1748858111281,
    "nanos": 1748858111281519000,
    "seconds": 1748858111
  },
  "hour": 19,
  "minute": 55,
  "month": 6,
  "second": 11,
  "time": "19:55:11",
  "year": 2025
}
```
