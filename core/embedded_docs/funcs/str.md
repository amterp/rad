# str

Converts any value to a string representation. Useful when you need to concatenate non-string values with `+`, though
interpolation (`"value: {x}"`) is generally preferred.

```rad
str(_var: any) -> str
```

```rad
str(42)        // -> "42"
str(3.14)      // -> "3.14"
str([1, 2])    // -> "[1, 2]"
str(true)      // -> "true"
```
