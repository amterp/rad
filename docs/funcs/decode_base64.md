# decode_base64

Decodes Base64 text back to original string.

## Signature

`decode_base64(_content: str, *, url_safe: bool = false, padding: bool = true) -> error|str`

## Examples

```rad
encoded = encode_base64("Hello World")
decoded = decode_base64(encoded)           // -> "Hello World"

// URL-safe decoding
url_encoded = encode_base64("test", url_safe=true)
decoded = decode_base64(url_encoded, url_safe=true)

// Handle errors - inside the block, `result` is the error
result = decode_base64("invalid base64!") catch:
    print_err("Decode failed: {result}")
    exit(1)

// Or fall back
result = decode_base64("invalid base64!") ?? ""
```

## Category

crypto

## Notes

**Parameters:**

| Parameter  | Type           | Description                                     |
|------------|----------------|-------------------------------------------------|
| `_content` | `str`          | Base64 text to decode                           |
| `url_safe` | `bool = false` | Expect URL-safe encoding (`-_` instead of `+/`) |
| `padding`  | `bool = true`  | Expect padding characters (`=`)                 |

Settings must match those used for encoding.
