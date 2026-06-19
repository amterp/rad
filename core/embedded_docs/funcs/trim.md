# trim

Strips all matching characters from both ends of a string. Preserves color attributes.

```rad
trim(_subject: str, _chars: str = " \t\n") -> str
```

```rad
trim("  hello  ")            // -> "hello"
trim("***hello***", "*")     // -> "hello"
trim("abcHELLOabc", "abc")   // -> "HELLO"
```
