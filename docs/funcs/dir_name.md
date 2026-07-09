# dir_name

Returns a path with its last element removed, i.e. the containing directory.

## Signature

`dir_name(_path: str) -> str`

## Examples

```rad
print(dir_name("/home/alice/notes.txt"))  // -> "/home/alice"
print(dir_name("src/core/funcs.go"))      // -> "src/core"
print(dir_name("standalone"))             // -> "."
```

## Category

io

## Notes

Pure string manipulation - the path does not need to exist.

Edge cases follow Unix `dirname` conventions: `dir_name("/")` returns `"/"`,
and `dir_name("")` returns `"."`. The result is lexically cleaned, e.g.
`dir_name("a/b/../c")` returns `"a"` after resolving the `..`.

See also [`base_name`](#base_name) and [`join_paths`](#join_paths).
