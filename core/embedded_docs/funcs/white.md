# white

Wraps its argument in the ANSI escape codes for white text.

```rad
white(_item: any) -> str
```

```rad
white("Hello")        // -> "Hello" wrapped in the white escape
white(42)             // -> "42" wrapped in the white escape
```

## See also

`black`, `plain`, `color_rgb`
