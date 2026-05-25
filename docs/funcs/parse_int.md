# parse_int

Parses a string to an integer.

## Signature

`parse_int(_str: str) -> int|error`

## Examples

```rad
parse_int("42")    // -> 42
parse_int("3.14")  // -> Error: invalid syntax
parse_int("abc")   // -> Error: invalid syntax
```

## Category

parsing
