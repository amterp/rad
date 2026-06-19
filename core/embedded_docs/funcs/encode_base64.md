# encode_base64

Encodes text to Base64 format.

```rad
encode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> str
```

```rad
encode_base64("Hello World")                      // -> "SGVsbG8gV29ybGQ="
encode_base64("Hello World", url_safe=true)       // -> URL-safe version
encode_base64("Hello World", padding=false)       // -> "SGVsbG8gV29ybGQ"
```

## Notes

**Parameters:**

| Parameter  | Type           | Description                                  |
| ---------- | -------------- | -------------------------------------------- |
| `_content` | `str`          | Text to encode                               |
| `url_safe` | `bool = false` | Replace `+/` with `-_` for URL-safe encoding |
| `padding`  | `bool = true`  | Include `=` padding characters               |

Use `url_safe=true` to replace `+/` with `-_` for URL-safe encoding. Use `padding=false` to omit `=` padding.
