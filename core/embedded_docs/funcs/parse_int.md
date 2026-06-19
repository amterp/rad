# parse_int

Parses a string to an integer.

```rad
parse_int(_str: str) -> int|error
```

```rad
parse_int("42")    // -> 42
parse_int("3.14")  // -> Error: invalid syntax
parse_int("abc")   // -> Error: invalid syntax
```
