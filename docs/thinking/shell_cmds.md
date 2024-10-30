# Shell Commands

Running shell commands *from RSL*.

## 2024-10-27

- Alternative approaches (including interpolation)
  1. `` out = `ls -l {dir}` ``
    - Might be undesirable to evaluate immediately on creation? Need some signal so that this can be treated as a string until then.
  2. `` out = $`ls -l {dir}` ``
    - Explicit signal to evaluate
    - xz does this.
  3. `` out = $"ls -l {dir}" ``
    - Avoids introducing additional symbol for 'strings'. However, might be annoying as `"` may conflict with the command that the user wants to write.
  4. `` out = shell("ls -l {dir}") ``
    - A bit cumbersome, and shares the issue about needing to escape `"` characters.
    - Also might make some syntax tricky / close us out from it, for example mandatory failure handling.
  5. `` out = shell(`ls -l {dir}`) ``
    - Seems like a more cumbersome version of (2). It might be more self-explanatory, but `$` is associated with shells and I think the writeability is worth it.

- All these are great for the standard 'run successfully and capture stdout' cases. I think we want to provide utilities beyond that:
  1. Capturing stderr
  2. Capturing error codes
  3. Responding to failure
     - Optionally? Or required?
  4. Think about security -- do we want to try and provide some controls against scripts that run arbitrary commands and wreak havoc?
     - Or should we just expect users to only run scripts they trust? Security might be cumbersome and worsen the experience for most people while they see no benefit?
     - Immediately, I don't think it's a major concern for our use case.
  5. Timeouts?
  6. Async?
     - *Potentially*. Not high prio.

- Let's imagine with go with syntax (2).
  - I think it's the most robust and positive tradeoff between writeability and readability.
  - Also provides a third way to write strings, which can be useful depending on what sort of contents you need and *where* you're writing.

### Capturing stderr, error codes, responding to failure

Optional handling:

```
// can just add additional vars to the capture.
// they're optional, can just do the one (out) if not interested in error handling.

out, err, code = $`ls -l {dir}`

// err is the stderr capture
// code is the error code int

if code != 0:
  print("Error: {err}")
  exit(code) 
```

The downside of this is that error handling is *opt-in*, whereas we might want to make it *opt-out*, to try and guide RSL developers into making reliable scripts. What syntax could we use for this?

Required failure block

```
stdout, err, code = $`ls -l {dir}`
failure:
  print("Error: {err}")
  exit(code)
```

Can add `unsafe` modifier to avoid requiring the 'failure' block. In this case, we'll just proceed if it fails, with the given vars.

```
stdout, err, code = unsafe $`ls -l {dir}`

// ... can proceed with the rest of the script
```

We want a keyword which will print the stderr and exit with the given code, though. What should it be?

... actually, I think that should just be the default behavior. Let's consider the following.

#### Opt-out failure handling

If code != 0, the following will print the err message, and exit with the given error code. No necessary handling/code required from RSL dev.

```
out, err, code = $`ls -l {dir}`

... rest of script
```

In the following case, if code != 0, then the failure block runs, allowing the RSL dev to handle it how they see fit.

```
out, err, code = $`ls -l {dir}`
failure:
  print("Error code {code}: {err}")
  exit(code)

... rest of script
```

If the RSL dev *does not want to handle the error*, and also *does not want the script to exit*, we have two options:

1. Comment fallthrough

```
out, err, code = $`ls -l {dir}`
failure:
  // no-op

... rest of script
```

2. `unsafe` keyword

```
out, err, code = unsafe $`ls -l {dir}`

... rest of script
```

It doesn't feel great to make (1) the canonical way to do this. (2) feels smoother and less hacky from a RSL perspective.

### Timeouts

```
out, err, code = $`ls -l {dir}` timeout 30  // 30 is a float, so you could also put e.g. 0.1 for 100 millis.
```

Does this syntax scale, though? Or get in the way if we want to add additional syntax around shell commands in the future?

I don't think it looks great with the `unsafe` syntax. Now you've got keywords on both sides of the shell expression.

Maybe don't worry for now. Maybe down the road we can revisit. Worst case, we do something like this

```
out, err, code = timeout 30 $`ls -l {dir}`
failure:
  // blah blah
timeout:
  // blah blah
```

### Async

Slightly alter `$` syntax?

```
future = $bg`long-running-task`
out, err, code = await(future)
```

### Extension of `failure` block to `rad` blocks

Naively extending:

```
rad url:
  fields Name, Age
failure:
  // handling
```

But, we could have *slightly* altered syntax for http codes:

```
rad url:
  fields Name, Age
http 4xx:
  // handling for user error
http 5xx:
  // handling for server-side error
```

All that said, I don't *love* the failure cases being at the same indentation as the rad. Is this better?

```
rad url:
  fields Name, Age
  http 4xx:
    // handling for user error
  http 5xx:
    // handling for server-side error
```

## 2024-10-30

```
// if this fails, we log the error, and proceed
unsafe $`ls -l`
```

```
// if this fails, we log the error, and end the script
$!`ls -l`
```

^ This syntax is not super self-explanatory (exclamation point is used as a non-null assertion in some languages, but it's not super widespread)
It may also be easy to miss that there's an exclamation point
But, maybe these shell stmts will be so common that people quickly get used to it and it'll be valuable for the syntax to be terse

```
// illegal syntax -- needs one of: unsafe, !, fail:, recover: (see below)
$`ls -l`
```

```
err = $`ls -l`
recover:
    // this block will run if the command failed.
    // after the block finishes, the script CONTINUES.
```

```
err = $`ls -l`
fail:
    // this block will run if the command failed.
    // after the block finishes, the script EXITS.
```

```
// invalid syntax -- only one of 'fail' or 'recover' allowed
err = $`ls -l`
fail:
    // fail block
recover:
    // recover block
```

Alternatively, share a preceding keyword:

```
err = $`ls -l`
catch recover:
    // same as recover above

// OR

err = $`ls -l`
catch fail:
    // same as fail above
```

^ I don't think 'catch' works here, as 'catch fail' doesn't naturally read as 'catch, then fail after this', instead it can be read as another way to say 'recover'. The two separate blocks is probably best.

---

Some thoughts, probably not but still here for the record:

```
err = $`ls -l`
retry 3, i:  // < i is a variable for the attempt #
    // this block will run if the command failed.
    // it will retry up to 3 times, rerunning the command and this block if it continues to fail
    // unclear what happens after 3 tries -- option to exit or continue the script? maybe we shouldn't offer this,
    // let users write their own retry algos. although, it'd be nice to solve for them if there're enough such use cases
```
