---
title: Error Handling
---

Things go wrong in scripts - users provide invalid input, files don't exist, network requests fail.
Generally, for smaller CLI scripts, it's okay if we just exit on the spot, and that's also Rad's default behavior.
However, if you wish to more gracefully handle these errors or attempt recovery, Rad gives you the tools to do so.

In Rad, **errors are values**. Functions that might fail return either a result or an error,
and you can decide how to handle them. This makes error handling explicit, predictable, and easy to reason about.

In this section, we'll explore:

- **Error propagation** - how errors bubble up by default (and why scripts exit)
- **The `catch:` block** - handle errors with full control (logging, reassignment, exit)
- **The `??` operator** - shorthand for fallback values on null or error
- **The `catch` operator** - like `??`, but only catches errors (null passes through)
- **Creating errors** - using `error()` in your own functions
- **Error type unions** - making fallible operations explicit with `T|error` type annotations

## Error Propagation

Let's start with a simple script that takes a user's age as input:

```rad linenums="1" title="File: age"
args:
    age_str str  # User's age as a string

age = parse_int(age_str)
print("You are {age} years old")
```

In reality, you'd instead declare the arg as an `int` and let Rad handle input validation,
but for the purposes of this guide, we write it as a `str`.

If we run this with a valid number, everything works:

```shell
rad age 25
```

<div class="result">
```
You are 25 years old
```
</div>

But what happens with invalid input?

```shell
rad age "not-a-number"
```

<div class="result">
```
Error at L4:7

age = parse_int(age_str)
      ^^^^^^^^^^^^^^^^^^ parse_int() failed to parse "not-a-number" (RAD20001)
```
</div>

The script **exits immediately** with an error code of 1 when `parse_int` encounters invalid input.
What's happening is that `parse_int` returned an `error` value, and since we're not handling it, it immediately gets propagated up.
Since we're at the root of the script and not nested within any other function call, this triggers a script exit on the spot.

### Nested Calls

Errors don't just propagate from built-in functions - they bubble up through your own function calls too. Here's an example:

```rad linenums="1"
fn calculate_discount(price_str: str) -> float|error:
    price = parse_float(price_str)  // Error starts here...
    return price * 0.1

fn process_order(item_price: str) -> str|error:
    discount = calculate_discount(item_price)  // ...propagates through here...
    return "Discount: {discount:.2} USD"

result = process_order("invalid")  // ...and exits the script here
print(result)
```

<div class="result">
```
Error at L2:13

      price = parse_float(price_str)  // Error starts here...
              ^^^^^^^^^^^^^^^^^^^^^^
              parse_float() failed to parse "invalid" (RAD20002)
```
</div>

The error originates in `parse_float`, propagates through `calculate_discount`, then through `process_order`, and finally exits at the top level. At any point in this chain, we could choose to handle the error instead of letting it propagate.

This sets up the question: how do we handle errors gracefully instead of crashing?

## Catch Blocks

The `catch:` block gives you full control over error handling. Attach it as a suffix to any expression that might error, and you can inspect the error, log it, provide a fallback value, or decide whether to exit.

### Basic Error Handling

Here's how to handle our age parsing example gracefully:

```rad linenums="1" title="File: age"
args:
    age_str str

age = parse_int(age_str) catch:
    print_err("Invalid age, falling back to 0: {age}")  // 'age' contains the error value
    age = 0  // Provide fallback

print("Age: {age}")
```

Now when we run it with invalid input:

```shell
rad age "not-a-number"
```

<div class="result">
```
Invalid age, falling back to 0: parse_int() failed to parse "not-a-number"
Age: 0
```
</div>

The script **continues running** with our fallback value. Inside the `catch:` block, the `age` variable contains the
error string, as returned by `parse_int`, which we can log or inspect. We then reassign `age` to a sensible default value of `0`.

To summarize:

- **Suffix** form: write `... catch:` directly after the error-able expression.
- **Binding**: the target variable is first bound to the error value; inside the block, interpolating it (e.g. `{age}`) prints the error’s message.
- **Control**: you can log, reassign a fallback, or exit(code).
- **Flow**: execution continues after the block unless you exit.

### Exiting on Errors

Sometimes you want to fail fast - handle the error just enough to log a helpful message, then exit:

```rad linenums="1" title="File: readconfig"
args:
    config_file str

config = read_file(config_file) catch:
    print_err("Failed to read config: {config}")
    exit(1)

print("Config loaded successfully")
// Continue processing config...
```

Running this with a non-existent file:

```shell
rad readconfig "missing.txt"
```

<div class="result">
```
Failed to read config: open missing.txt: no such file or directory
```
</div>

This example is not much better than the default error propagation and exit, but you can imagine providing
more useful guidance to users in a more detailed error message.

### Ignoring Errors with `pass`

Sometimes you want to ignore errors entirely - the operation might fail, but that's perfectly fine and requires no action:

```rad linenums="1"
// Custom fn to clean up temp file if it exists
delete_path(temp_file) catch:
    pass  // File already doesn't exist, that's fine

// Continue with the rest of the script...
```

Here, `pass` does nothing - it's a way to explicitly say "I know this might error, but I don't care." This is useful for cleanup operations where the failure itself is harmless.

## The `??` Operator

For simple cases where you just want a default value without any logging or conditional logic, the `??` operator provides a concise shorthand:

```rad
age = parse_int(age_str) ?? 0
timeout = parse_int(get_env("TIMEOUT")) ?? 30
max_retries = parse_int(config["retries"]) ?? 5
```

`??` fires when the left side is **null or an error**, making it a null-coalescing operator:

```rad
name = user["name"] ?? "anonymous"   // handles both missing keys and null values
config = read_file(config_path) ?? get_default_config()
```

The right side uses **lazy evaluation** - it's only evaluated if the left side is null or an error. This means you can safely call functions on the right without worrying about unnecessary work.

This makes `??` useful for safely drilling into nested data. If any key along the way is missing or null, the whole expression falls back:

```rad
name = response.user.profile.display_name ?? "anonymous"
```

## The `catch` Operator

The `catch` operator is similar to `??`, but **only catches errors** - null values pass through unchanged:

```rad
age = parse_int(age_str) catch 0      // error -> 0, but null stays null
data = parse_json(raw_input) catch {}  // parse failure -> empty map
```

This is useful when you want to handle errors but need to preserve null as a meaningful value:

```rad
m = {"key": null}
m["key"] ?? "fallback"     // -> "fallback" (?? treats null as missing)
m["key"] catch "fallback"  // -> null       (catch lets null through)
```

Like `??`, `catch` supports lazy evaluation and chaining:

```rad
result = risky_call() catch fallback_call() catch "final default"
```

!!! note "Not to be confused with `catch:` blocks"

    The `catch` operator is an inline expression. The `catch:` block (with a colon) is a statement-level construct covered [above](#catch-blocks) that gives you full control, including logging and conditional exit.

### Comparing `??`, `catch`, and `catch:`

These three give you different levels of control:

```rad
// ?? - fallback on null or error
age = parse_int(age_str) ?? 0

// catch - fallback on error only (null passes through)
age = parse_int(age_str) catch 0

// catch: block - full control (logging, conditional handling)
age = parse_int(age_str) catch:
    print_err("Invalid age '{age_str}': {age}")
    age = 0
```

!!! tip "When to use which"

    Use `??` when you want a default for both null and error cases.
    Use `catch` when you only want to handle errors and null is a valid value.
    Use `catch:` when you need to log, inspect, or conditionally handle the error.

## Creating Your Own Errors

When writing your own functions, you can return errors using the `error(str)` function.
If you're using type annotations, then functions that may return errors should reflect that in its return type: `T|error`.

```rad linenums="1"
fn validate_port(port: int) -> int|error:
    if port < 1 or port > 65535:
        return error("Port must be between 1-65535, got {port}")
    return port

fn start_server(port_str: str) -> void:
    port = parse_int(port_str) ?? 8080

    validated_port = validate_port(port) catch:
        print_err("Invalid port: {validated_port}")
        exit(1)

    print("Starting server on port {validated_port}")

start_server("99999")
```

<div class="result">
```
Invalid port: Port must be between 1-65535, got 99999
```
</div>

Our custom error message provides clear feedback about what went wrong.
By returning `int|error`, the type signature tells you three things:

1. This function normally returns an `int`
2. It might return an `error` instead
3. Callers should consider handling the error case (otherwise let it propagate)

This pattern is used throughout Rad's built-in functions:

- `parse_int(str) -> int|error`
- `parse_float(str) -> float|error`
- `read_file(path) -> error|{ "size_bytes": int, "content": str }`
- `round(num, decimals) -> error|int|float`

The error union makes your code self-documenting - anyone reading your function signature knows immediately that it can fail.

!!! info "More on Union Types"

    We covered union types in detail in an earlier section: [Type Annotations](./type-annotations.md#union-types).
    Error unions are just one application of Rad's union type system.

## Summary

Rad's error handling model gives you the tools to write robust scripts that handle failures gracefully:

- **Errors are values** that **propagate** by default, unless handled
- **Scripts exit** if errors propagate up to the root of the script
- **`catch:` blocks** provide full error handling control:
    - Variable contains the error string inside the block
    - You can log errors, provide fallbacks, or call `exit()`
    - Execution continues unless you explicitly exit
- **`??` operator** provides concise fallbacks on null or error
    - Use for simple cases without logging
    - Right side only evaluated if left side is null or errors
- **`catch` operator** provides fallbacks on error only
    - Null values pass through unchanged
    - Useful when null is a meaningful value you want to preserve
- **Create errors** with `error("message")` in your own functions
- **Type unions** (`T|error`) make fallible operations explicit in function signatures

## Next

CLI scripts and the shell go hand in hand, and Rad offers first-class support for invoking shell commands and handling its output.
We explore this in the next section: [Shell Commands](./shell-commands.md).
