# split_lines

Splits a string by line endings. Handles all common styles: `\n` (Unix), `\r\n` (Windows), and `\r` (legacy Mac).

## Signature

`split_lines(_val: str) -> str[]`

## Examples

```rad
"a\nb\nc".split_lines()          // -> ["a", "b", "c"]
content = read_file("data.txt").content
for line in content.split_lines():
    print(line)
```

## Category

strings

## Notes

Use this instead of `split("\n")` when processing text that may come from different platforms.


Trailing line endings are stripped - `"a\nb\n".split_lines()` returns `["a", "b"]`, not `["a", "b", ""]`.
