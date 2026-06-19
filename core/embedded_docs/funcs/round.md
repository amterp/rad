# round

Rounds a number to the specified decimal precision.

```rad
round(_num: float, _decimals: int = 0) -> error|int|float
```

```rad
round(3.14159)           // -> 3 (integer)
round(3.14159, 2)        // -> 3.14 (float)
round(2.7)               // -> 3 (integer)
round(3.14, -1)          // -> Error: precision must be non-negative
```

## Notes

**Parameters:**

| Parameter   | Type      | Description                                     |
| ----------- | --------- | ----------------------------------------------- |
| `_num`      | `float`   | Number to round                                 |
| `_decimals` | `int = 0` | Number of decimal places (must be non-negative) |

With precision 0, returns an integer. With precision > 0, returns a float. Precision must be non-negative.
