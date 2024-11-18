# Rest Args

Capturing "the rest" of provided args as an array to be used in the script for e.g. passthroughs.

---

```rsl
args:
    name string
    age a float
    rest
```

```
script Alice -a 30 -t quiet
```

```
name = Alice
age = 30
rest = ["-t", "quiet"]
```

What should the following do? `-a 30` is recognized by the script.

```
script Alice -t quiet -a 30
```

Option 1:

If there's a `rest` arg in the script, as soon as we see an unknown arg, we just start putting those into the 'rest' array.
So in this particular case, it will throw because `age` is unspecified. 
If age had a default '20', it would succeed with the following values 

```
name = Alice
age = 20
rest = ["-t", "quiet", "-a", "30"]
```

Option 2:

Skip unknown args/flags, use recognized flags, put the unrecognized into an array.

```
name = Alice
age = 30
rest = ["-t", "quiet"]
```

Option 2 makes it easier to treat it as a implementation detail that the script wraps another tool. For option 1, you need to be aware of which flags the script itself takes, versus the ones that are acceptable to get passed downstream.

Though maybe the latter is actually a virtue. It's more likely to result in correct usage. But arguably confusing if you expect to be able to tag a flag onto the back that the script recognizes, but then this becomes part of the passthrough? What about if that tagged-on flag is `-h`?

---
