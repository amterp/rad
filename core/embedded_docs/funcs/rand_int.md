# rand_int

Returns a random integer in a specified range.

```rad
rand_int(_arg1: int = 9223372036854775807, _arg2: int?) -> int
```

```rad
rand_int(10)        // -> Random int from 0-9
rand_int(5, 15)     // -> Random int from 5-14
rand_int(10, 5)     // -> Error: min (10) must be less than max (5)
```

## Notes

With one argument, returns random int from 0 to `_arg1` (exclusive). With two arguments, returns random int from `_arg1`
to `_arg2` (exclusive). Min must be less than max.
