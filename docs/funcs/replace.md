# replace

Replaces every occurrence of a literal substring. Pass `regex=true` to treat `_find` as a regex pattern and enable capture-group references in `_replace`. Does not preserve string color attributes.

## Signature

`replace(_original: str, _find: str, _replace: str, *, regex: bool = false) -> str`

## Parameters

- `_original` (`str`): The string to search.
- `_find` (`str`): The text to find. Literal by default, a regex pattern when `regex=true`.
- `_replace` (`str`): The replacement. Fully literal by default; supports `$` group references when `regex=true`.
- `regex` (`bool = false`): Treat `_find` as a regex pattern.

## Examples

```rad
replace("hello world", "world", "Rad")           // -> "hello Rad"
replace("1.2.3", ".", "-")                       // -> "1-2-3"
replace("cost: $5", "$5", "$10")                 // -> "cost: $10"
replace("abc123def", "\\d+", "XXX", regex=true)  // -> "abcXXXdef"
replace("Name: Charlie Brown", "Charlie (.*)", "Alice $1", regex=true) // -> "Name: Alice Brown"
```

## Category

strings

## Notes

`_find` is matched literally unless `regex=true`, so metacharacters like `.` and `(` need no
escaping in the default case. In literal mode `_replace` is inserted verbatim, `$` included -
safe for text relayed from elsewhere.

When `regex=true`, a pattern that fails to compile is an error. Patterns use Go's RE2 dialect,
which has no backreferences or lookaround.

**Group references in `_replace` (`regex=true` only):**

| Written        | Expands to                                             |
|----------------|--------------------------------------------------------|
| `$0`           | The whole match                                        |
| `$N`           | Capture group `N`; `N` is the longest run of digits    |
| `${N}`         | Capture group `N`, explicitly delimited                |
| `$$`           | A literal `$`                                          |
| Anything else  | Left as written                                        |

A reference to a group that does not exist is left as written rather than expanding to an
empty string, so stray `$` in your replacement text survives intact.

Because `$N` consumes every following digit, use `${N}` when a literal digit has to follow a
group reference: `${1}0` inserts group 1 followed by `0`, while `$10` means group 10.

`${N}` collides with Rad's string interpolation, which would read `{1}` as an expression to
interpolate. Write the replacement as a raw string or escape the brace:

```rad
replace("Name: abc", "a(b)c", r"${1}0", regex=true)   // -> "Name: b0"
replace("Name: abc", "a(b)c", "$\{1}0", regex=true)   // -> "Name: b0"
```

## See also

`split`, `matches`
