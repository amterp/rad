# underline

Wraps its argument in the ANSI escape codes for underlined text.

```rad
underline(_item: any) -> str
```

```rad
underline("Hello")        // -> "Hello" wrapped in the underline escape
underline(42)             // -> "42" wrapped in the underline escape
```

## See also

`bold`, `italic`
