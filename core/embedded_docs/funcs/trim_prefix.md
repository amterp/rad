# trim_prefix

Removes a literal prefix from the start of a string (once). Preserves color attributes.

```rad
trim_prefix(_subject: str, _prefix: str) -> str
```

```rad
trim_prefix("hello world", "hello ")  // -> "world"
trim_prefix("aaabbb", "a")            // -> "aabbb" (one 'a' removed)
trim_prefix("test", "x")              // -> "test" (no match)
```
