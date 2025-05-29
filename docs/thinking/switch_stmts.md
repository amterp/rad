# Switch Statements

Switch stmt approach:

```rad
title, url = switch endpoint:
    case "cars": "Cars", "{base}/automobiles"
    case "books": "Books", "{base}/reading?type=books"
```

Map approach:

```rad
opts = [
    {
        "keys": ["cars"],
        "values": ["Cars", "{base}/automobiles"],
    },
    {
        "keys": ["books"],
        "values": ["Books", "{base}/reading?type=books"],
    },
]
title, url = pick_kv(endpoint, opts)
```

A lot more verbose, have to remember 'opts' structure.

---

There's some overlap between file resources and switch statements. I think we should allow users to avoid resource files and to embed them as switch statements in their scripts of they so desire. Resource files can be useful for if the resource file changes a lot or benefits from automation in being updated.

So, let's consider going ahead with the switch statement approach.

---

What if you want switch *blocks*?

```rad
switch endpoint:
    case bloop:
        this is
        some rsl code
    case bar, foo:
        this is additional code
        in a block
    default:
        default handling here
```

but then we need to make sure this is compatible with switch *expressions* that return something in e.g. assignments.

---

Simple switch with blocks:

```
switch foo:
    case "alice":
        print("Alice!")
    case "bob":
        print("Bob!")
    default:
        print("Who?!")
```

Allowed to have single statement on same line as case:

```
switch foo:
    case "alice": print("Alice!")
    case "bob": print("Bob!")
    default: print("Who?!")
```

Can assign:

```
text = switch foo:
    case "alice": "Alice!"
    case "bob": "Bob!"
    default: "Who?!"
```

Can assign with blocks:

```
text = switch foo:
    case "alice":
        "Alice!"
    case "bob":
        "Bob!"
    default:
        "Who?!"
```

Last line will be evaluated for the return value of the switch expr. Can also assign multiple

```
text, num = switch foo:
    case "alice":
        print("Alice case")
        "Alice!", 5
    case "bob":
        print("Bob case")
        "Bob!", 6
    default:
        print("Default case")
        "Who?!", 0
```

Can mix "block" and in-line returns:

```
text, num = switch foo:
    case "alice": "Alice!", 5
    case "bob": "Bob!", 6
    default:
        print("Unknown!")
        exit(1)
```
