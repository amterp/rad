# Macros

## 2025-05-03

We want two things. They may both be solved by the same mechanism, or by two separate mechanisms.

1. A way for the script to reference some piece of information statically available from rad, particularly metadata.

For example, main use case is referencing the script name in the file header.

```
---
Does something.
Examples:
  @script alice 30
  @script alice -a 30
---
```

Here, without rad running any of the script, it can recognize the `@script` and replace it with the script's actual name
when generating the usage string.

2. Supplying information *to* rad, statically.

The main use case at the moment is stash id. Something like this perhaps?

```
---
Does something.
@script_id = J3nSdEa1v5T
---
```

This latter case is a little tenuous. `@script` is nice because it's basically one token.
The `@script_id` usage is not, though. There's an equal. There's another token (string?).
What if this is written on the same line as other text? It's just kinda weird. Feels like a different syntax is needed.

---

So the `@` syntax works well when you're *invoking* information, maybe less when you're *providing* it.
Either we find a different syntax for the latter, or one which works better for both.

To reiterate, the goal is to provide and retrieve this information *statically*. The interpreter should not need to
actually *run* any of the script.

## 2025-05-04

Maybe we just try something. We can adjust syntax later.

```
---
Does something.
Example: @script alice 30
Blah blah @@script << escaped? 
^ That has to be on its own line. Anything after `=` is interpreted as the value.
---
@script_id = J3nSdEa1v5T
---
```
