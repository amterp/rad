# Things to Revisit

- regex strings need to be more explicitly opted-into (e.g. for split func)
- `quiet` being how we suppress announcements for shell commands
- body should be a named arg in http methods?
- one `http` func for all methods instead of some having their own?
- `{}` vs. `${}` for interpolation
- Ruby-style % syntax or bash heredocs/herestrings?
- Division by 0 errors or returns nan?
- `del a[1], a[1]` on a list - currently deletes two different items, but should perhaps be atomic and "delete the same one"

- get rid of go-like if stmts

```
if a; b:
  c
```

is not meaningfully shorter than

```
a
if b:
  c
```

and 
only confusing to those unfamiliar.

---

- `excludes` constraint being bidirectional

```
args:
  a int
  b int

  a excludes b
```

- ^ could argue that defining just `a` should be allowed, but *not* just defining `b`, as we did not say `b excludes a`.
- In that particular example, not allowing just `b` implies you *need* to define `a`, but again, that precludes `b`, so you can just never define `b`.
- However, maybe we can come up with an alternative configuration where it *does* make sense.

Another thought: perhaps `a excludes b` on bool args should mean "b cannot be true if a is".

Example where you want mutually exclusive bools:
```
args:
    verbose v bool
    quiet q bool
```

---
