# print_err

Behaves like [`print`](#print) but outputs to stderr instead of stdout.

## Signature

`print_err(*_items: any, *, sep: str = " ", end: str = "\n") -> void`

## Examples

```rad
print_err("failed to load config")     // -> writes to stderr
print_err("error:", err.msg)           // -> "error: <msg>" to stderr
```

## Category

io
