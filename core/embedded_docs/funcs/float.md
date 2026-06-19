# float

Converts a value to a float. Does not work on strings - use `parse_float` for string parsing.

```rad
float(_var: any) -> float|error
```

```rad
float(42)      // -> 42.0
float(true)    // -> 1.0
float(false)   // -> 0.0  
float("3.14")  // -> Error: cannot convert string
```
