# index_of

Finds the index of a target value within a string or list. Returns `null` if not found.

## Signature

`index_of(_subject: str|list, _target: any, *, n: int = 0, start: int = 0) -> int?`

## Examples

```rad
// String search
"hello world hello".index_of("hello")           // -> 0
"hello world hello".index_of("hello", n=1)      // -> 12
"hello world hello".index_of("hello", n=-1)     // -> 12
"hello".index_of("xyz")                         // -> null
"hello".index_of("xyz") ?? (-1)                 // -> -1
"hello".index_of("")                             // -> null (empty target)

// List search
["a", "b", "c", "b", "a"].index_of("b")        // -> 1
["a", "b", "c", "b", "a"].index_of("b", n=-1)  // -> 3
[1, 2, 3].index_of(99)                          // -> null
```

## Category

strings

## Notes

**Parameters:**

| Parameter  | Type        | Description                                           |
|------------|-------------|-------------------------------------------------------|
| `_subject` | `str\|list` | The string or list to search within                   |
| `_target`  | `any`       | The value to search for                               |
| `n`        | `int = 0`   | Which occurrence to find (0=first, 1=second, -1=last) |
| `start`    | `int = 0`   | Position to start searching from                      |
