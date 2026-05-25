# gen_fid

Generates a random [flex ID](https://github.com/amterp/flexid) (fid) - a time-ordered, URL-safe identifier.

## Signature

`gen_fid(*, alphabet: str?, tick_size_ms: int?, num_random_chars: int?) -> error|str`

## Examples

```rad
gen_fid()                                    // -> "1a2b3c4d5e"
gen_fid(alphabet="0123456789")               // -> "1234567890"
gen_fid(num_random_chars=3)                  // -> "1a2b3c"
```

## Category

crypto

## Notes

**Parameters:**

| Parameter          | Type                       | Description                            |
|--------------------|----------------------------|----------------------------------------|
| `alphabet`         | `str? = "[0-9][A-Z][a-z]"` | Characters to use (base-62 by default) |
| `tick_size_ms`     | `int? = 1`                 | Time precision in milliseconds         |
| `num_random_chars` | `int? = 6`                 | Number of random characters to append  |

Defaults: `alphabet` is base-62 (`[0-9][A-Z][a-z]`), `tick_size_ms` is 1ms, `num_random_chars` is 6.
