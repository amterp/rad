# matches

Tests whether `_str` matches the regular-expression `_pattern`. By default the pattern must match the whole string; pass `partial=true` to match any substring. Returns an `error` when the pattern is malformed.

## Signature

`matches(_str: str, _pattern: str, *, partial: bool = false) -> bool|error`

## Examples

```rad
matches("hello", "h.+o")               // -> true
matches("hello world", "world")        // -> false (default is full-string match)
matches("hello world", "world", partial=true)  // -> true
matches("abc", "(")                    // -> error: invalid regex

```

## Category

strings

## Notes

Unlike `split` and `replace`, `matches` has no `regex` parameter and always treats `_pattern`
as a regex. A literal `matches` would just be `==` (full match) or a substring test (partial),
neither of which needs a function.

## See also

`replace`, `split`
