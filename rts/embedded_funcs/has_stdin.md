# has_stdin

Checks if stdin is piped to the script.

## Signature

`has_stdin() -> bool`

## Examples

```rad
has_stdin()                     // -> true (if piped)
has_stdin()                     // -> false (if not piped)
if has_stdin():
  content = read_stdin()        // Conditional read
```

## Category

system
