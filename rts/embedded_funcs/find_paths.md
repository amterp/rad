# find_paths

Returns a list of all paths under a directory.

## Signature

`find_paths(_path: str, *, depth: int = -1, relative: ["target", "cwd", "absolute"] = "target") -> error|str[]`

## Examples

```rad
// Find all files in directory
paths = find_paths("src/")
for path in paths:
    print(path)  // -> "file1.txt", "subdir/file2.txt", etc.

// Limit depth
paths = find_paths("src/", depth=1)  // -> Only direct children

// Get absolute paths
paths = find_paths("src/", relative="absolute")
```

## Category

io

## Notes

**Parameters:**

| Parameter  | Type                                       | Description                            |
|------------|--------------------------------------------|----------------------------------------|
| `_path`    | `str`                                      | Directory to search                    |
| `depth`    | `int = -1`                                 | Max depth to search (-1 for unlimited) |
| `relative` | `["target", "cwd", "absolute"] = "target"` | How to format returned paths           |

- `"target"` - Relative to input path (default)
- `"cwd"` - Relative to current directory
- `"absolute"` - Full absolute paths

A leading `~` in `_path` is expanded to your home directory.
