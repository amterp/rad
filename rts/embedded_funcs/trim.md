# trim

Strips all matching characters from both ends of a string. Preserves color attributes.

## Signature

`trim(_subject: str, _chars: str = " \t\n") -> str`

## Examples

```rad
trim("  hello  ")            // -> "hello"
trim("***hello***", "*")     // -> "hello"
trim("abcHELLOabc", "abc")   // -> "HELLO"
```

## Category

strings
