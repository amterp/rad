# read_file

Reads the contents of a file.

## Signature

`read_file(_path: str, *, mode: ["text", "bytes"] = "text") -> error|{ "size_bytes": int, "content": str|int[] }`

## Examples

```rad
// Read text file
result = read_file("config.txt")
if result.success:
    content = result.content  // -> string
    
// Read binary file
result = read_file("image.png", mode="bytes")
if result.success:
    bytes = result.content    // -> list[int]
    
// Handle errors
result = read_file("missing.txt")
if not result.success:
    print("Error:", result.error)
```

## Category

io

## Notes

**Parameters:**

| Parameter | Type                         | Description                     |
|-----------|------------------------------|---------------------------------|
| `_path`   | `str`                        | Path to the file to read        |
| `mode`    | `["text", "bytes"] = "text"` | Read as UTF-8 text or raw bytes |

In text mode, decodes as UTF-8 and returns a string. In bytes mode, returns a list of integers.

**Return map contains:**

- `size_bytes: int` - File size in bytes
- `content: str|list[int]` - File contents (type depends on mode)

**Examples:**
