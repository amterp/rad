# parse_float

Parses a string to a float.

```rad
parse_float(_str: str) -> float|error
```

```rad
parse_float("3.14")  // -> 3.14
parse_float("42")    // -> 42.0
parse_float("abc")   // -> Error: invalid syntax
```
