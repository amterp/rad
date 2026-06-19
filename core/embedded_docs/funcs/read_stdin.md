# read_stdin

Reads all data from stdin.

```rad
read_stdin() -> str?|error
```

```rad
read_stdin()                  // -> "piped content" (if piped)
read_stdin()                  // -> null (if not piped)
read_stdin()                  // -> Error 20026 if read fails
content = read_stdin()
lines = content.split_lines() // Process stdin line-by-line
```
