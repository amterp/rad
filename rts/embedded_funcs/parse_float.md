# parse_float

Parses a string to a float.

## Signature

`parse_float(_str: str) -> float|error`

## Examples

```rad
parse_float("3.14")  // -> 3.14
parse_float("42")    // -> 42.0
parse_float("abc")   // -> Error: invalid syntax
```

## Category

parsing
