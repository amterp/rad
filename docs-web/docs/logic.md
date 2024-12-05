---
title: Boolean Logic
---

## Truthy / Falsy

RSL supports truthy/falsy logic.

For those unfamiliar, this means that, instead of writing the following (as an example):

```rsl
if len(myList) > 0:
    print("My list has elements!")
```

you can write

```rsl
if myList:
    print("My list has elements!")
```

Essentially, you can use any type as a condition, and it will resolve to true or false depending on the value.

See below for which values for each type will resolve to false. All other values will resolve to true.

| Type   | Falsy |
|--------|-------|
| string | `""`  |
| int    | `0`   |
| float  | `0.0` |
| list   | `[]`  |
| map    | `{}`  |

Note that a string which is all whitespace e.g. `" "` is truthy.
