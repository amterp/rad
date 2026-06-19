# filter

Applies a predicate function to filter elements of a list or map. Keeps only elements where the function returns true.

```rad
filter(_coll: map|list, _fn: fn(any) -> bool | fn(any, any) -> bool) -> map|list
```

```rad
filter([1, 2, 3, 4], fn(x) x % 2 == 0)      // -> [2, 4]
filter({"a": 1, "b": 2}, fn(k, v) v > 1)    // -> {"b": 2}
```

## Notes

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)`.
