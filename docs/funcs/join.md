# join

Joins a list into a string with separator, prefix, and suffix.

## Signature

`join(_list: list, sep: str = "", prefix: str = "", suffix: str = "") -> str`

## Examples

```rad
join([1, 2, 3], sep=", ")           // -> "1, 2, 3"
join(["a", "b"], prefix="[", suffix="]")  // -> "[ab]"
join(["x", "y", "z"], sep="-", prefix="(", suffix=")")  // -> "(x-y-z)"
```

## Category

lists
