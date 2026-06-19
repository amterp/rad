# len

Returns the number of elements in a string, list, or map. For
strings this is the rune count (not byte count), so unicode characters
contribute one each.

```rad
len(_val: str|list|map) -> int
```

```rad
len("hello")              // -> 5
len([1, 2, 3])            // -> 3
len({"a": 1, "b": 2})     // -> 2
len("héllo")              // -> 5 (rune count, not byte count)
```

## Parameters

- `_val` (`str|list|map`): the collection to measure.

## See also

`sort`, `keys`, `values`
