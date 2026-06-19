# strikethrough

Wraps its argument in the ANSI escape codes for strikethrough text. Renders with a line through it on terminals that support the attribute.

```rad
strikethrough(_item: any) -> str
```

```rad
strikethrough("Hello")        // -> "Hello" wrapped in the strikethrough escape
strikethrough(42)             // -> "42" wrapped in the strikethrough escape
```

## See also

`underline`, `dim`
