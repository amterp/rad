# sort

Returns a new sorted list (or string with characters sorted). The
input is not mutated. With `reverse=true`, sorts in descending order.
Multiple lists / strings can be passed - they're sorted in lockstep
using the first as the key.

```rad
sort(_primary: list|str, *_others: list|str, *, reverse: bool = false) -> list|str
```

```rad
sort([3, 1, 2])             // -> [1, 2, 3]
sort([3, 1, 2], reverse=true)  // -> [3, 2, 1]
sort("dcba")                // -> "abcd"

ages = [30, 25, 28]
names = ["alice", "bob", "carol"]
sort(ages, names)           // -> [25, 28, 30]
                            //    names is now ["bob", "carol", "alice"]
```

## Parameters

- `_primary` (`list|str`): the sequence to sort by.
- `_others` (`variadic `list|str`): additional sequences that get
- `reverse` (`bool`, keyword-only, default `false`): when true,

## See also

`len`, `reverse`
