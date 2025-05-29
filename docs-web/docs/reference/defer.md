---
title: Defer & Errdefer
---

- `defer` and `errdefer` run in LIFO order, each kind being part of the same one queue.
- If there are several defer statements, and one fails, further defer statements will still attempt to run.
- Rad's error code will become an error if the main script succeeded but a defer statement failed.
- errdefers will not get triggered if the main script succeeded but a `defer` or `errdefer` statement failed.

## `defer`

```rad title="defer Example"
defer:
    print(1)
    print(2)
defer:
    print(3)
    print(4)
print("Hello!")
```

```title="defer Example Output"
Hello!
3
4
1
2
```

## `errdefer`

```rad title="errdefer Example 1"
defer:
    print(1)
    print(2)
errdefer:
    print(3)
    print(4)
defer:
    print(5)
    print(6)
errdefer:
    print(7)
    print(8)
print("Hello!")
exit(0)  // successful script run
```

```title="errdefer Example 1 Output"
Hello!
5
6
1
2
```

```rad title="errdefer Example 2"
defer:
    print(1)
    print(2)
errdefer:
    print(3)
    print(4)
defer:
    print(5)
    print(6)
errdefer:
    print(7)
    print(8)
print("Hello!")
exit(1)  // perceived as error!
```

```title="errdefer Example 2 Output"
Hello!
7
8
5
6
3
4
1
2
```
