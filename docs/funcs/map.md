# map

Applies a function to every element of a list or entry of a map.

## Signature

`map(_coll: map|list, _fn: fn(any) -> any | fn(any, any) -> any) -> map|list`

## Examples

```rad
map([1, 2, 3], fn(x) x * 2)              // -> [2, 4, 6]
map({"a": 1, "b": 2}, fn(k, v) v * 10)   // -> {"a": 10, "b": 20}
```

## Category

lists

## Notes

For lists, function receives `fn(value)`. For maps, function receives `fn(key, value)`.
