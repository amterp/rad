# int

Converts a value to an integer. Does not work on strings - use [`parse_int`](#parse_int) for string parsing.

## Signature

`int(_var: any) -> int|error`

## Examples

```rad
int(3.14)     // -> 3
int(true)     // -> 1
int(false)    // -> 0
int("42")     // -> Error: cannot convert string
```

## Category

parsing
