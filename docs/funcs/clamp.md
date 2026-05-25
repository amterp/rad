# clamp

Constrains a value between minimum and maximum bounds.

## Signature

`clamp(val: int|float, min: int|float, max: int|float) -> error|int|float`

## Examples

```rad
clamp(25, 20, 30)    // -> 25
clamp(10, 20, 30)    // -> 20
clamp(40, 20, 30)    // -> 30
clamp(5, 1.0, 10)    // -> 5.0 (float because 1.0 is float)
clamp(15, 30, 20)    // -> Error: min must be <= max
```

## Category

math

## Notes

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `val`     | `int | float`      | Value to constrain |
| `min`     | `int | float`      | Minimum bound      |
| `max`     | `int | float`      | Maximum bound      |

Returns `val` if between min and max, otherwise returns the nearest bound. Min must be ≤ max.
The return type preserves the input type: returns `int` if all inputs are integers, `float` if any input is a float.
