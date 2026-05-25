# type_of

Returns the type of a value as a string.

## Signature

`type_of(_var: any) -> ["int", "str", "list", "map", "float", "bool", "null", "error", "function"]`

## Examples

```rad
type_of("hi")            // -> "str"
type_of([2])             // -> "list"
type_of(42)              // -> "int"
type_of(3.14)            // -> "float"
type_of({"a": 1})        // -> "map"
type_of(true)            // -> "bool"
type_of(null)            // -> "null"
type_of(fn() 1)          // -> "function"
// Builtins that may fail return an `error` value:
// type_of(parse_int("xx")) // -> "error"
```

## Category

system
