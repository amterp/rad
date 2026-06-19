# sleep

Pauses execution for the specified duration.

```rad
sleep(_duration: int|float|str, *, title: str?) -> void
```

```rad
sleep(2.5)              // -> Sleep for 2.5 seconds
sleep("1h30m")          // -> Sleep for 1 hour 30 minutes
sleep("500ms")          // -> Sleep for 500 milliseconds
sleep("1d12h")          // -> Sleep for 1 day 12 hours
sleep(5, title="Waiting...") // -> Prints "Waiting..." then sleeps 5 seconds
```

## Notes

Integer and float values are treated as seconds. String values support duration format like "2h45m", "1.5s", "500ms".
Spaces are allowed in duration strings (e.g. `"5m 30s"`).
If `title` is provided, it's printed before sleeping.

**Duration string suffixes:**

| Suffix       | Description  |
| ------------ | ------------ |
| `d`          | Days         |
| `h`          | Hours        |
| `m`          | Minutes      |
| `s`          | Seconds      |
| `ms`         | Milliseconds |
| `us` or `µs` | Microseconds |
| `ns`         | Nanoseconds  |
