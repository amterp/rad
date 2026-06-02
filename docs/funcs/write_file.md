# write_file

Writes content to a file. Creates the file if it doesn't exist.

## Signature

`write_file(_path: str, _content: str, *, append: bool = false) -> error|{ "bytes_written": int, "path": str }`

## Examples

```rad
// Write new file
result = write_file("output.txt", "Hello world")
print("Wrote", result.bytes_written, "bytes")

// Append to existing file
write_file("log.txt", "\nNew entry", append=true)

// Error handling
result, err = write_file("/readonly/file.txt", "data")
if err:
    print("Write failed:", err.msg)
```

## Category

io

## Notes

**Parameters:**

| Parameter  | Type           | Description                                       |
|------------|----------------|---------------------------------------------------|
| `_path`    | `str`          | Path where to write the file                      |
| `_content` | `str`          | Content to write                                  |
| `append`   | `bool = false` | Append to existing content instead of overwriting |

By default overwrites the file. Use `append=true` to append to existing content.

A leading `~` in `_path` is expanded to your home directory.

**Return map contains:**

- `bytes_written: int` - Number of bytes written
- `path: str` - Full path to the written file

**Examples:**
