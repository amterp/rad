# split

Splits a string on a literal separator. Pass `regex=true` to treat `_sep` as a regex pattern instead. Does not preserve string color attributes.

## Signature

`split(_val: str, _sep: str, *, limit: int?, regex: bool = false) -> str[]`

## Parameters

- `_val` (`str`): The string to split.
- `_sep` (`str`): The separator. Literal text by default, a regex pattern when `regex=true`.
- `limit` (`int?`): Caps the number of splits performed. Must be >= 1.
- `regex` (`bool = false`): Treat `_sep` as a regex pattern.

## Examples

```rad
split("a,b,c", ",")                    // -> ["a", "b", "c"]
split("1.2.3", ".")                    // -> ["1", "2", "3"]
split("word1  word2", "\\s+", regex=true) // -> ["word1", "word2"]
split("abc123def", "\\d+", regex=true) // -> ["abc", "def"]
split("key=val=ue", "=", limit=1)      // -> ["key", "val=ue"]
split("a,b,c,d", ",", limit=2)         // -> ["a", "b", "c,d"]
```

## Category

strings

## Notes

`_sep` is matched literally unless `regex=true`, so metacharacters like `.` and `(` need no
escaping in the default case.

When `regex=true`, a pattern that fails to compile is an error. Patterns use Go's RE2 dialect,
which has no backreferences or lookaround.

When `limit` is provided, it caps the number of splits performed. The final element contains
the unsplit remainder. `limit` must be >= 1.

An empty `_sep` splits between every character.

## See also

`split_lines`, `replace`, `matches`
