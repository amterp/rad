# bold

Wraps its argument in the ANSI escape codes for bold text.

```rad
bold(_item: any) -> str
```

```rad
bold("Hello")        // -> "Hello" wrapped in the bold escape
bold(42)             // -> "42" wrapped in the bold escape
```

## See also

`dim`, `italic`, `underline`
