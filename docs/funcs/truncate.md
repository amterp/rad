# truncate

Truncates a string to a maximum length, adding an ellipsis if truncated. Requires length of at least 1.

## Signature

`truncate(_str: str, _len: int) -> error|str`

## Examples

```rad
truncate("hello world", 8)   // -> "hello w…"
truncate("short", 10)        // -> "short" (no truncation needed)
truncate("test", 0)          // -> Error: Requires at least 1
```

## Category

strings
