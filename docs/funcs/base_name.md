# base_name

Returns the last element of a path.

## Signature

`base_name(_path: str) -> str`

## Examples

```rad
print(base_name("/home/alice/notes.txt"))  // -> "notes.txt"
print(base_name("src/core/"))              // -> "core" (trailing slashes are ignored)
print(base_name("standalone"))             // -> "standalone"
```

## Category

io

## Notes

Pure string manipulation - the path does not need to exist, unlike
[`get_path`](#get_path), which only reports `base_name` for existing paths.

Edge cases follow Unix `basename` conventions: `base_name("/")` returns `"/"`,
and `base_name("")` returns `"."`.

See also [`dir_name`](#dir_name) and [`join_paths`](#join_paths).
