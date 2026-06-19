# dim

Wraps its argument in the ANSI escape codes for dimmed text - the reverse of bold, useful for de-emphasising less important output.

```rad
dim(_item: any) -> str
```

```rad
dim("Hello")        // -> "Hello" wrapped in the dim escape
dim(42)             // -> "42" wrapped in the dim escape
```

## See also

`bold`, `italic`
