---
title: Boolean Logic
---

## Truthy / Falsy

RSL supports truthy/falsy logic.

For those unfamiliar, this means that, instead of writing the following (as an example):

```rad
if len(my_list) > 0:
    print("My list has elements!")
```

you can write

```rad
if my_list:
    print("My list has elements!")
```

Essentially, you can use any type as a condition, and it will resolve to true or false depending on the value.

The following table shows which values return false for each type. **All other values resolve to true.**

| Type   | Falsy | Description   |
|--------|-------|---------------|
| string | `""`  | Empty strings |
| int    | `0`   | Zero          |
| float  | `0.0` | Zero          |
| list   | `[]`  | Empty lists   |
| map    | `{}`  | Empty maps    |

!!! note ""

    Note that a string which is all whitespace e.g. `" "` is truthy.
