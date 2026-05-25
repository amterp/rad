# trim_suffix

Removes a literal suffix from the end of a string (once). Preserves color attributes.

## Signature

`trim_suffix(_subject: str, _suffix: str) -> str`

## Examples

```rad
trim_suffix("hello world", " world")  // -> "hello"
trim_suffix("aaabbb", "b")            // -> "aaabb" (one 'b' removed)
trim_suffix("test", "x")              // -> "test" (no match)
```

## Category

strings
