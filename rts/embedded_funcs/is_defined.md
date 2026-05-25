# is_defined

Checks if a variable with the given name exists in the current scope.

## Signature

`is_defined(_var: str) -> bool`

## Examples

```rad
name = "Alice"
is_defined("name")     // -> true
is_defined("age")      // -> false
```

## Category

system
