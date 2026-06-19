# has_stdin

Checks if stdin is piped to the script.

```rad
has_stdin() -> bool
```

```rad
has_stdin()                     // -> true (if piped)
has_stdin()                     // -> false (if not piped)
if has_stdin():
  content = read_stdin()        // Conditional read
```
