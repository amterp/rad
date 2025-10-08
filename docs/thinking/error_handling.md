# Error Handling

## 2025-10-08

Right now, anything that can fail will default to erroring and exiting the script immediately (or rather, propagating the error),
unless the dev opts into error handling.

We currently have two error kinds of error handling.

1. `catch`

```
result = catch parse_int(text)
if type_of(result) == "error":
    ...
```

2. shell invocations

```
$`curl '{url}'`
recover:
    ... // run this code, then continue with script because we want to recover
    
$`curl '{url}'`
fail:
    ... // run this code, then exit the script because we want to fail the script

$!`curl '{url}'` // critical command that exits on the spot if failed

unsafe $`curl '{url}'`  // non-critical command, explicitly opting out of recover/fail handling
```

1) is extremely primitive, verbose, and error prone.
2) is somewhat neat, but very unique syntax (learning curve) and the rules can be hard to remember

Ideally, 1) and 2) are actually *merged* into the same syntax. They're doing the same thing -- we're doing some operation that can fail, then deciding how (and if) to handle it!

Let's focus first on the function case, since it's a little simpler.

A) Error block

```
result = catch parse_int(text) error:
    // ... handle error
// continue with successful case

// rewritten more purely:

result = catch parse_int(text) error:
    print_err("Failed to parse {text}")
    exit(1)
```

actually we can probably remove the `catch` keyword entirely there and shuffle it to be more Zig-like:

```
result = parse_int(text) catch e:
    print_err("Failed to parse {text}: {e}")
    exit(1)
```

I wonder if this syntax would always work though, with all expressions...

This is quite concise, and unlike Type Switch below, it handles the most common error handling pattern, so that gives this a big bonus.

B) Type switch

```
result = catch parse_int(text)
switch result:
    case error -> // single line operation; OR
    case error:
        // multiline handle

// rewritten more purely:

result = catch parse_int(text)
switch result:
    case error:
        print_err("Failed to parse {text}")
        exit(1)
```

This solution is not mutually exclusive from A), especially since switching on type generally could be useful in a loosely typed language like Rad, especially when a variable has several possible types.

Tangentially, do we need the `case` keyword? I've considered getting rid of it before... Can we?

Back to the switch statement. What about inlining `result`:

```
switch catch parse_int(text):
    case error -> ...
    case int -> ...
```

`catch` arguably a bit dumb here... I don't think we can get rid of it though.
TBC

```
result = try parse_int(text):
    error -> ...
    int -> ...
```

This syntax is interesting, but what if you just wanna proceed with the int, outside of this `try` construct? The `int` branch seems silly.

We could make a semantic whereby, any unhandled types are implied to do `<type> -> yield result` ? So:

```
result = try parse_int(text):
    error:
        print_err("Failed to parse {text}")
        exit(1)

// make it to here only if `result` is not an error, it's an int in this case 
```

This is not bad. I don't know if `try` is right though, this construct also just feels useful for switching on types, again. For example, imagine a `calculate()` function which returns either an int, string, or list.

```
result = try calculate():
    int -> ...
    string -> ...
    list -> ...
```

`try` isn't quite right... It's not like it's gonna fail. `switch` ?

```
result = switch calculate():
    int -> ...
    string -> ...
    list -> ...
```

Again, back to this just being a generally useful concept, we can do this for other reasons as well. But how does it feel for our error case? Let's say `calculate` also could return `error`.

...

Thinking more about it, I don't think there's a satisfying syntax using `switch` which concisely let's users handle only errors, and just assign the result for non-errors.

```
result = switch calculate():
    error:
        print_err("Failed to parse {text}")
        exit(1)
```

This feels weird I think. We're switching on the result of `calculate()`, doing something if it's an error, but is it obvious what happens with all the other results?
Is it implied `calculate()` can only return errors? Maybe I'm overthinking it, maybe it's clear enough and it's actually a quite elegant solution, which
gives us error handling for a 'type switch' syntax we'll probably implement anyway?

What if we don't want to handle the error though? Just exit?

```
result = calculate()
```

Well okay that's simple enough. But what if we wanted to switch to handle the int case?

```
result = switch calculate():
    int -> result * 2
```

Tangent but this is a bit weird, we don't have a temporary var for referring to the output of `calculate()`, so `result` is both the temporary and the final var we assign to... Is this confusing?
It's easy to come up with a rule that `result` is initially equal to result of `calculate()` and then gets re-assigned to the result of a branch (if matched). Hmm.

Anyway, this block of code is for when we don't want to handle errors. So I suppose a rule is "unhandled error returns will throw". Seems sane.
And if you just wanna actually get the error value, you could do `error -> yield result`, but also if we keep the `catch` syntax, you can still just do `catch calculate()` if you don't need to do any other switching.

---

Another tangent, but I'm interested in exploring this syntax (demonstrative only):

```
result = parse_int(text) or 0
result = parse_int(text) ?? 0  // alternatively
```

Right now, the first line is legal syntax but won't work because parse_int() still throws an error. If we did `catch parse_int(text)`, then we `or` evaluates, but the error is truthy and so `result` just becomes the error.
I don't think I want to change the behavior of `or` in this case; that's working as intended.
`??` could have a different meaning though -- reject only errors or nulls maybe?

---

Okay trying to draw some conclusions for this thinking session.

1) I like the `result = parse_int(text) catch e:` syntax. You could even drop the ` e` if you don't need the error (tho maybe we should actually disallow that, in an effort to get devs printing their errors).
2) `??` seems interesting, let's explore. Claude actually suggests `catch` here: `port = parse_int(env) catch get_default_port()`, interesting 🤔 Don't think we do that.
3) Type switching is maybe still useful, but not urgent, we can delay.
4) `catch parse_int(text)` might actually be unnecessary. We can probably replace it with the suffix variation entirely.
5) We can apply the catch suffix approach to shell commands probably. Drop the `$!`, all `$` invocations are implicitly critical. Can error handle:

```
$`curl '{url}'`

$`curl '{url}'` catch e:
    print_err("curl failed! {e}")
    exit(1) // or don't, to recover!
```

And we can get rid of the `unsafe` keyword. You could equivalently do

```
$`curl '{url}'` catch e:
    pass
```
