# trim_right

Strips all matching characters from the end of a string. Preserves color attributes.

```rad
trim_right(_subject: str, _chars: str = " \t\n") -> str
```

```rad
trim_right("  hello  ")         // -> "  hello"
trim_right("***hello***", "*")  // -> "***hello"
trim_right("aaabbb", "b")       // -> "aaa"
```
