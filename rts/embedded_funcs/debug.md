# debug

Behaves like [`print`](#print) but only outputs when debug mode is enabled via `--debug` flag.

## Signature

`debug(*_items: any, *, sep: str = " ", end: str = "\n") -> void`

## Examples

```rad
debug("entering loop")             // -> nothing unless --debug is on
debug("x =", x, "y =", y)          // -> debug-only diagnostics
```

## Category

io
