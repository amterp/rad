# trim_left

Strips all matching characters from the start of a string. Preserves color attributes.

## Signature

`trim_left(_subject: str, _chars: str = " \t\n") -> str`

## Examples

```rad
trim_left("  hello  ")          // -> "hello  "
trim_left("***hello***", "*")   // -> "hello***"
trim_left("aaabbb", "a")        // -> "bbb"
```

## Category

strings
