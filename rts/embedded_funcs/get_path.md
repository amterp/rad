# get_path

Gets information about a file or directory path.

## Signature

`get_path(_path: str) -> { "exists": bool, "full_path": str, "base_name"?: str, "permissions"?: str, "type"?: str, "size_bytes"?: int, "modified_millis"?: int, "accessed_millis"?: int }`

## Examples

```rad
info = get_path("config.txt")
if info.exists:
    print("File size:", info.size_bytes, "bytes")
    print("Type:", info.type)
else:
    print("File not found")

// Working with timestamps using parse_epoch()
info = get_path("data.json")
if info.exists:
    mtime = info.modified_millis.parse_epoch()
    print("Last modified:", mtime.date, mtime.time)
```

## Category

system

## Notes

**Always returns:**

- `exists: bool` - Whether the path exists
- `full_path: str` - Absolute path

**When path exists, also returns:**

- `base_name?: str` - File/directory name
- `permissions?: str` - Permission string (e.g., "rwxr-xr-x")
- `type?: str` - Either "file" or "dir"
- `size_bytes?: int` - File size (only for files)
- `modified_millis?: int` - Modification time as epoch milliseconds
- `accessed_millis?: int` - Access time as epoch milliseconds (Unix/macOS only)

**Examples:**
