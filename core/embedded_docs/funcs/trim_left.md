# trim_left

Strips all matching characters from the start of a string. Preserves color attributes.

```rad
trim_left(_subject: str, _chars: str = " \t\n") -> str
```

```rad
trim_left("  hello  ")          // -> "hello  "
trim_left("***hello***", "*")   // -> "hello***"
trim_left("aaabbb", "a")        // -> "bbb"
```
