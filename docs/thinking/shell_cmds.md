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

## 2025-10-13

We're revisiting our shell commands syntax, especially off the back of our latest thinking on the subject of [error handling](./error_handling.md#2025-10-08).
TLDR we're thinking of having this syntax for error handling:

```
result = parse_int(text)  // no error handling, so if parse_int returns an error, we immediately propagate that error up
result = parse_int(text) ?? 0  // if parse_int returns an error, we fall back to 0. 0 is a lazy expression, so could be a function

result = parse_int(text) catch e:  // newline and code block MUST follow `catch e:`
    print_err("Failed to parse: {e}")  // variable after `catch` becomes the error (probably, see option C below)
    exit(1)
```

Seems clear, but there are some outstanding questions.
We'd like to use this sort of syntax for our shell commands as well, so maybe thinking through that will help
us find answers to said questions.

Question: What if we don't exit from the catch block? What is `result`?

```
result = parse_int(text) catch e:
    print_err("Failed to parse: {e}")

// ... script continues here, we didn't exit. What is 'result'?
```

Option A is that we leave `result` undefined.
Option B is that we require catch blocks to *return* or *yield* something.

```
result = parse_int(text) catch e:
    print_err("Failed to parse: {e}")
    yield 0

// `result` is now 0, if parse_int failed.
```

This is sort of like expanding `??` out into a block. It's also not unlike our `switch` syntax, where a case can immediately give a value
`case 0 -> "hi"`, or be a more complex block:

```
result = switch x:
    case 0:
        print("zero!")
        yield "hi"
```

Option C is that we ditch the `e` error variable, and still assign `result` to whatever `parse_int` returned, including the error itself:

```
result = parse_int(text) catch:
    print_err("Failed to parse: {result}")  // prints `result` which is the error, if we're in this block
    exit(1)
```

Maybe there are other options I'm failing to think of right now.

Okay, so that's the direction we're thinking about for error handling. Now, shell commands.

Currently, we have an `unsafe` keyword, we have the concept of regular (`$`) and critical (`$!`) shell commands.
We have the `recover` and `fail` blocks that are required for regular invocations (but not critical).

This is a bit complex. It's unusual syntax and contributes to Rad's learning curve, and even I mess it up at times.
Also, in practice, I've found that I am using `$!` syntax 95% of the time -- I want scripts to fail on the spot if
a shell cmd fails.

On top of this, we return values from shell invocations in a special way. You can either get 0, 1, 2, or 3 variables out of a shell command.

```
$!`grep hello`  // no vars, prints everything (stdout/stderr) straight to terminal
code = $!`grep hello` // captures just the exit code as an int, stdout/stderr straight to terminal
code, stdout = $!`grep hello` // stdout is now captured, doesn't go to terminal, goes into buffer and captured as string variable. stderr still straight to terminal
code, stdout, stderr = $!`grep hello` // both stdout and stderr now captured, and nothing goes to terminal
```

A couple of reasons why we do it this way:
1) Often you don't want to use those variables in your script, so you leave them undefined.
2) For a lot of commands, you don't wanna interfere with the output at all, you wanna let it go straight to the user's terminal (e.g. for interactive CLIs, etc). Capturing them would break them.
3) Compact syntax! It's actually very convenient to use.

Downsides:

1) Can't capture stderr without capturing stdout.

Somewhat relatedly, that's also why the code is first -- so you can capture the exit code, without having to capture stdout/stderr

2) If you don't care about e.g. the exit code, you have to use a place holder to get to later outputs:

```
_, stdout = $!`grep hello` // bad
_, _, stderr = $!`grep hello` // worse
```

This doesn't bother me a ton tbh, but I've gotten feedback from one user that they thought this was ugly.

So, in an ideal world, our syntax allows us to keep the first three benefits I mentioned, while avoiding these two downsides.

If we can't think of a syntax to tick all our boxes by the way, it's possible we have a "simple" and "advanced" syntax, for those cases that need it.

Anyway, back to error handling.

While trying to meet the above constraints, we'd also like to make the syntax work with our new error handling ideas, but it's potentially tricky.

Let's get easy changes out of the way:

- We get rid of the `$` / `$!` alternatives -- just `$` now, and all invocations are critical, UNLESS the user opts into error handling (`??` or `catch`).
- No `unsafe` keyword, no `fail` or `recover` blocks.

So, the immediate tricky thing, with our existing syntax, is that shell invocations don't just return one value, they return 3 (sort of).

For this `parse_int(text) ?? 0`, it's clear what `??` has to check -- the single return value of the expression to its left. Similar for `catch`.

A semantic we've imposed in Rad is that functions (and shell commands) don't actually ever return more than one value. Instead, they may return a list, containing several values.

So, if we have a function `my_func()`, which returns three values, these two statements are equivalent.

```
[a, b, c] = my_func()  // destructuring
a, b, c = my_func()    // sugar; exactly equivalant to above destructuring
```

I think we should think of shell invocations the same way -- they return a list of between 0 and 3 elements, and you can destructure it. The only difference
unique to shell invocations (that functions don't have) is that shell commands, in their Rad implementation, can see how many vars they're being assigned to,
and so we can alter its capturing behavior based on that (the syntax doesn't allow shell invocations in expressions to e.g. function calls etc, only as statements or assignments).

Okay. So what do these alternatives this do?

```
$`grep hello` ?? 0

$`grep hello` catch e:
    pass
```

For `??`, maybe its values on the right should match the assign? So in that example, we should actually allow no value, as a shorthand for saying "just continue if error"?
And if there are assignments, the number of assignments must match?

```
$`grep hello` ??
code = $`grep hello` ?? 1
code, stdout, stderr = $`grep hello` ?? 1, "", ""
```

Not sure I like any of that, except maybe the 0-case for `??` ? Tho the syntax might look confusing to people unfamiliar.

What about `catch`? What is `e` exactly? Maybe similar, where number of vars should match?

```
$`grep hello` catch: // nothing captured
    pass

$`grep hello` catch code, stdout, stderr:  // can do 1, 2, or all 3 of these
    pass
    
// becomes weird here though, we're forced to always capture stderr for the sake of the error case, despite
// us not trying to assign stderr for the non-error case
code, stdout = $`grep hello` catch code, stdout, stderr:
    pass
```

Again, not sure I like any of that.

I don't have time as of writing this entry to keep going so I'll leave it with some brief notes to think more about:

- A complex access object for accessing shell output? But how know as time of shell invocation what it needs to contain, and what to capture? Shell syntax somehow allow that?
- Claude suggested this i.e. no error variable, we maintain the assignment. Should this apply to non-shell error handling case to, for consistency?

```
  code, stdout = $`grep hello file` catch:
      // Variables ARE assigned to the actual (failed) values
      // No separate 'e' needed - the error info IS code/stdout/stderr
      print_err("Command failed with code {code}")
      print_err("Output was: {stdout}")
      // Can exit or continue

  The key: assigned variables contain the error-state values. No magic e variable because the error information is already captured in what you assigned.
```

- It also suggested this `??` approach, I actually like the explicit list

```
  Option A: Require explicit list:
  code, stdout = $`grep hello file` ?? [1, ""]

  Option B: Only allow ?? for single captures; use catch for multiple:
  code, stdout = $`grep hello file` catch:
      yield 1, ""  // parallel to switch syntax
```

Separate from shell invocations, it also argues for this with error handling on functions:

```
I think Option C makes the most sense - result IS the error. No separate e.

result = parse_int(text) catch:
    print_err("Parse failed: {result}")  // result is the error
    result = 0  // explicit assign as fallback
```

Which is consistent with shell commands.

ChatGPT had this to contribute, which is not bad. On the subject of shell invocation return values:

```
Named capture pattern (advanced, optional)

Add an alternative LHS that’s named, order-independent, so you can capture exactly what you want with no placeholders:

{code} = $`grep hello`                 // only code
{stderr} = $`grep hello`               // only stderr
{stdout: out} = $`grep hello`          // alias binding
{code, stderr} = $`grep hello`         // both, order-free
{stdout, stderr} = $`grep hello`       // no code
```

I don't know if I agree with that `{}` syntax specifically, but the idea is intriguing. I asked Claude about it's thoughts
in this space, and it suggested similar

```
If Rad supports (or will support) map destructuring, you could allow both positional and named access:

// Current positional syntax (keep this - it's great for common cases)
code = $`cmd`
code, stdout = $`cmd`
code, stdout, stderr = $`cmd`

// New: Named destructuring for selective capture
{stderr} = $`cmd`  // Just stderr, stdout → terminal
{stdout, stderr} = $`cmd`  // Skip code entirely
{code, stderr} = $`cmd`  // Skip stdout
```

Actually, maybe we don't even need the fancy `{}` syntax? Maybe as a default, we can just *always* detect this case? If any
of your vars are named `stdout` or `stderr`, we'll assign them accordingly? After all, you'll never want `stdout` to
get assigned to a variable named `stderr`. I think this could work!!

TLDR tentative conclusions:

- Continue to return 0 to 3 vars from shell invocations, but introduce named assignments, enabled by default
- Do not have a `e` variable for `catch`, just assign the output of the function/shell invocation always, and use them in the catch block
- Keep catch blocks as a series of statements -- the block doesn't yield anything, assigned vars can be re-assigned to fallback values
- Allow `??` to return several values, but via a list `[ ]`. This includes for shell: `code, stderr = $'grep hello' ?? [1, 'failed']`
- I *think* empty `??` should *not* be allowed for simple skips. It's not self-explanatory syntax, though maybe my mind will change. For now, just do a `catch` block and `pass`.
- edit: for longer expressions guarded by a `catch`, what is the assigned variable equal to? I guess whatever the error was, regardless of where it came from? see below

ONE LAST TANGENT THOUGHT

We could allow shell invocations to be used as expressions. What if we allowed syntax like this?

```
branch = $`git branch --show-current`.stdout.trim()
```

The way it would work, is that, if shell invocations are used in an expression (not as a statement), then we capture code, stdout, and stderr, and return a
map containing those three things as keys. `.stdout` is using key access syntax in Rad (completely standard to all maps), and we're trimming it all in one
go. A big thing to decide on, is if the shell invocation is critical when used this way, or not.

If it is, then we're being pretty strict, but users can do this:

```
branch = $`git branch --show-current`.stdout.trim() catch:
    branch = "undefined"
```

still, since `catch` is so loosely binding. Though, it's a little unclear what `branch` is equal to, inside the catch block?
Actually this is also true for non-shell invocation examples, I'll edit the above conclusions list with this (though it might be
simple for that case). Maybe `branch` is technically just equal to the error code? And it's on the user to not read the `branch`
variable in the error block since it's sorta nonsense? or just make it null? Actually, claude suggests it could be equal to
the map returned by the shell invocation, maybe that'd work. if you're nesting several shell cmds into one `catch` guard, then
perhaps we could also add `.cmd` as a property in this map, so you can print in your error which command actually failed.
That said, I don't love that your map gets a weird name e.g. `branch`. Maybe this is an argument for having the `e` variable,
and we could just make it a duplicate of the left side? For multi-returning functions, it could be a list, or for shell invocations,
it could be the map. We could also allow destructuring, at least the list case:

```
my_func() catch [var1, var2]:
    pass
// this would be more controversial in the shell invocation case, as again, we'd
// be forced to capture for the sake of errors, probably.
```

If the above example is *not* critical, then `stdout` is probably just a blank string? So `branch` ends up blank? I think this is worse, as
it makes it harder to detect what actually happened in your operation.
