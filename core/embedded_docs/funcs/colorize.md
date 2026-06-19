# colorize

Assigns consistent colors to values from a set of possible values. The same value always gets the same color within the
same set.

```rad
colorize(_val: any, _enum: any[], *, skip_if_single: bool = false) -> str
```

```rad
names = ["Alice", "Bob", "Charlie"]
colorize("Alice", names)     // -> "Alice" (in consistent color)
colorize("Bob", names)       // -> "Bob" (in different consistent color)

// In rad blocks
names = ["Alice", "Bob", "Charlie", "David"]
rad:
    fields names
    names:
        map fn(n) colorize(n, names)
```

## Notes

**Parameters:**

| Parameter        | Type           | Description                                    |
| ---------------- | -------------- | ---------------------------------------------- |
| `_val`           | `any`          | Value to colorize                              |
| `_enum`          | `any[]`        | Set of possible values for consistent coloring |
| `skip_if_single` | `bool = false` | Don't colorize if only one value in set        |

Useful for automatically coloring table data or distinguishing values in lists.
