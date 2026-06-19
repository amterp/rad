# italic

Wraps its argument in the ANSI escape codes for italic text. Not every terminal renders italics; some show inverse or coloured text instead.

```rad
italic(_item: any) -> str
```

```rad
italic("Hello")        // -> "Hello" wrapped in the italic escape
italic(42)             // -> "42" wrapped in the italic escape
```

## See also

`bold`, `underline`
