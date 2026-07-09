# mkdir

Creates a directory, including any missing parent directories.

## Signature

`mkdir(_path: str) -> error|{ "path": str, "created": bool }`

## Examples

```rad
res = mkdir("output/reports")
print(res.created)  // -> true (false if it already existed)

// Idempotent - safe to call when the directory already exists
res = mkdir("output/reports")
print(res.created)  // -> false

// path is the input after ~ expansion, with forward slashes
res = mkdir("~/backups")
print(res.path)     // -> e.g. "/home/alice/backups"

// Errors are catchable
res = mkdir("/root/forbidden") catch:
    print("Could not create:", res)
```

## Category

io

## Notes

Behaves like `mkdir -p`: creates all missing parents, and succeeds (with
`created: false`) if the directory already exists. Directories are created
with `0755` permissions.

Errors if the path exists but is a file, or if creation fails (e.g. missing
permissions).

A leading `~` in `_path` is expanded to your home directory. The returned
`path` is the input after that expansion, normalized to forward slashes on
all platforms - it is not resolved to an absolute path.

Counterpart to [`delete_path`](#delete_path).
