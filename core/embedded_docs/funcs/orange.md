# orange

Wraps its argument in the ANSI escape codes for orange text. Rendered via the closest 256-colour palette entry on terminals that don't support 24-bit colour.

```rad
orange(_item: any) -> str
```

```rad
orange("Hello")        // -> "Hello" wrapped in the orange escape
orange(42)             // -> "42" wrapped in the orange escape
```

## See also

`yellow`, `red`, `color_rgb`
