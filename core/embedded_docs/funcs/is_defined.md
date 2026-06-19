# is_defined

Checks if a variable with the given name exists in the current scope.

```rad
is_defined(_var: str) -> bool
```

```rad
name = "Alice"
is_defined("name")     // -> true
is_defined("age")      // -> false
```
