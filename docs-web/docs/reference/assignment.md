---
title: Assignment
---

Generally speaking, multi-assignments are only legal for switch expressions, or single operations (e.g. functions) that return multiple values.

## Legal Assignments

```rad
a = 1
a, b = pick_from_resoure(...)
a, b = switch ...
a, b = parse_int(text)

myMap["key"] = 2
myList[1] = 3
```

## Illegal Assignments

```rad
a, b = 1, 2
myMap["key"], myMap["key2"] = 2, 3
myList[1], myList[2] = 3, 4
```
