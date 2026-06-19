# pink

Wraps its argument in the ANSI escape codes for pink text. Rendered via the closest 256-colour palette entry on terminals that don't support 24-bit colour.

```rad
pink(_item: any) -> str
```

```rad
pink("Hello")        // -> "Hello" wrapped in the pink escape
pink(42)             // -> "42" wrapped in the pink escape
```

## See also

`magenta`, `red`, `color_rgb`
