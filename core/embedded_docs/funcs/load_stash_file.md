# load_stash_file

Loads a file from the script's stash directory, creating it with default content if it doesn't exist.

```rad
load_stash_file(_path: str, _default: str = "") -> error|{ "full_path": str, "created": bool, "content"?: str }
```

```rad
result = load_stash_file("config.txt", "default config")
if result.success:
    if result.created:
        print("Created new config file")
    content = result.content
```

## Notes

**Return map contains:**

- `full_path: str` - Full path to the file
- `created: bool` - Whether the file was just created
- `content?: str` - File contents (if successfully loaded)
