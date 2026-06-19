# hash

Generates a hash of the input text using various algorithms.

```rad
hash(_val: str, algo: ["sha1", "sha256", "sha512", "md5"] = "sha1") -> str
```

```rad
hash("hello world")                    // -> "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
hash("hello world", algo="sha256")     // -> "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
hash("sensitive data", algo="sha512")  // -> Long SHA-512 hash
```

## Notes

**Parameters:**

| Parameter | Type                                           | Description              |
| --------- | ---------------------------------------------- | ------------------------ |
| `_val`    | `str`                                          | Text to hash             |
| `algo`    | `["sha1", "sha256", "sha512", "md5"] = "sha1"` | Hashing algorithm to use |

The default `sha1` is **not cryptographically secure**. Use `sha256` or `sha512` for security.
