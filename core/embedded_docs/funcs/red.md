# red

Wraps its argument in the ANSI escape codes for red text.

```rad
red(_item: any) -> str
```

```rad
red("Hello")        // -> "Hello" wrapped in the red escape
red(42)             // -> "42" wrapped in the red escape
```

## See also

`green`, `yellow`, `color_rgb`
