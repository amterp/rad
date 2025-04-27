# Custom Functions & Lambdas

Some thinking over on [stashes.md](./stashes.md).

TLDR of the syntax I concluded on there:

```rsl
normalize = fn(x) x.trim().lower()
 
normalize(mystring)

normalize = fn(x):
    out = x.trim().lower()
    return out

provide = fn() 5
provide()  // returns 5

multiply = fn(x, y) x * y

mylist.map(fn(x) x.upper())
mylist.map(upper)  // technically, it'd need to redefine all my built-ins as function vars, so they can be passed this way
```

---

Related topic: [imports.md](./imports.md).
