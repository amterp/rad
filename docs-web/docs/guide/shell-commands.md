---
title: Shell Commands
---

The shell offers a wide range of utilities and is essential for CLI scripting - from file operations to invoking
installed programs like git, make, or docker.

Rad has rich built-in functionality (`http_get`, `read_file`, `write_file`, etc.), but sometimes you need to invoke
system tools or installed programs. Rad makes this safe and ergonomic through first-class shell command support.

## Invoking Commands

Shell commands are invoked by prefixing a string with `$`:

```rad
$`ls -la`
```

You can also pre-define the command as a string variable:

```rad
cmd = `ls -la`
$cmd
```

By default, the stdout/stderr will be printed directly to the user's terminal as if they had invoked it directly themselves.

!!! tip "Prefer backticks for shell command strings"

    Shell commands often use 'single' and "double" quotes, so backticks minimize delimiter conflicts.
    However, you can use any string delimiter.

## Capturing Output

Shell commands return three values: **exit code**, **stdout**, and **stderr**. You can capture anywhere from zero to all three of these values, depending on what you need.

### Capture Modes

There are four levels of capture:

**1. No capture - output goes to terminal**

When you don't assign any variables, all output goes to the terminal:

```rad
$`ls -la`
```

**2. Capture exit code only**

Assign to one variable to capture just the exit code:

```rad
code = $`make test`
```

The exit code is captured as an `int`, but stdout and stderr still go to the terminal.

**3. Capture exit code + stdout**

Assign to two variables to capture the exit code and stdout:

```rad
code, stdout = $`git show 0dd21e6`
```

The exit code and stdout are captured as an `int` and `str` respectively. Stderr still goes to the terminal.
**Important:** When you capture stdout, it doesn't print to the terminal - it's redirected to your variable.

**4. Capture all three**

Assign to three variables to capture everything:

```rad
code, stdout, stderr = $`npm install`
```

All three values are captured. Nothing is printed to the terminal automatically.

### Named Assignment

Rad supports a special form of assignment when working with shell commands. When **all** your variables are named
exactly `code`, `stdout`, or `stderr`, then assignment happens by name rather than by position.
This means the order doesn't matter:

```rad
// Named assignment - order independent
stdout, code = $`echo hi`           // code=0, stdout="hi\n"
stderr = $`bad-command`             // Just capture stderr
code, stderr = $`make format`       // code=1, stderr=""
stderr, stdout, code = $`ls`        // All three, any order
```

This improves readability - you can capture exactly what you need with clear, self-documenting variable names.

**The rule:** If ALL variables use exactly `code`, `stdout`, or `stderr`, assignment is by name. Otherwise, it's positional:

```rad
// Positional - 'output' isn't a special name
code, output = $`echo hi`           // output = stdout (by position)
exit_code, out, err = $`ls`         // Assigned in order
```

This lets you write clear code like `stderr = $cmd` instead of `_, _, stderr = $cmd`.

!!! tip "Silencing outputs"

    You can use `_` to ignore specific outputs: `code, _ = $cmd` captures the code and ignores stdout.
    For silent execution, capture everything: `_, _, _ = $cmd` - nothing will print to the terminal.

## Error Handling

Now that you understand how to capture output, let's talk about error handling.

When a shell command exits with a non-zero exit code, it triggers error propagation - just like functions that return errors.
This means you can handle potential failures using `catch:` blocks:

```rad
// Handle errors with catch block
$`make build` catch:
    print_err("Build failed!".red())
    exit(1)

// Or ignore failures
$`make build` catch:
    pass  // Continue on failure
```

You can **combine capturing with error handling**. When the `catch:` block runs, your variables are already
assigned their actual values, so you can inspect them:

```rad
// Capture the exit code AND handle errors
code = $`make test` catch:
    print_err("Command failed to run. Error code {code}")
    exit(1)

print("Tests passed!")
```

This works with any capture pattern:

```rad
code, stdout = $`git tag --list` catch:
    print_err("Failed to get tags")
    exit(1)

version = stdout.trim()
```

This uses the same error model covered in [Error Handling](./error-handling.md) - errors propagate by default, so you need `catch:` blocks to handle them.

## String Interpolation

You can build commands dynamically using string interpolation:

```rad linenums="1"
args:
    version str
    message str

// Interpolate variables into commands
$`git tag v{version}` catch:
    print_err("Failed to create tag")
    exit(1)

$`git commit -m "{message}"` catch:
    print_err("Commit failed")
    exit(1)
```

This is particularly useful for constructing commands based on script arguments or other runtime values.

## Modifiers

Rad provides two modifiers that can be applied to shell commands.

### The `quiet` Modifier

By default, Rad announces each shell command with a ⚡️ prefix. For example, this command:

```rad
$`touch hello.txt` catch:
    print_err("Failed to create file")
    exit(1)
```

Shows in the terminal:

<div class="result">
```
⚡️ touch hello.txt
```
</div>

To suppress this announcement, use the `quiet` modifier:

```rad
quiet $`touch hello.txt` catch:
    print_err("Failed to create file")
    exit(1)
```

<div class="result">
```
(no output - unless there's an error)
```
</div>

This is useful for scripts that run many commands or when you want minimal output.

### The `confirm` Modifier

The `confirm` modifier prompts the user before running a command:

```rad
confirm $`rm -rf node_modules`
```

This is particularly useful for destructive operations.

## Practical Examples

Let's look at some real-world patterns that combine these features.

### Development Workflow

Here's a script inspired by a typical development workflow:

```rad
---
Validates code, checks git status, and optionally pushes changes.
---
args:
    push p bool  # Push changes after validation

// Run validation steps
steps = ["go mod tidy", "make format", "make build", "make test"]

for step in steps:
    $step catch:
        print_err("❌ {step} failed".red())
        exit(1)
    print("✅ {step} passed".green())

if push:
    // Check for uncommitted changes
    stdout = $`git status --porcelain` catch:
        print_err("Failed to check git status")
        exit(1)

    if stdout.trim() != "":
        print_err("Working directory has uncommitted changes!")
        print_err("Commit your changes before pushing.")
        exit(1)

    // Get current branch and push
    stdout = $`git branch --show-current` catch:
        print_err("Failed to get current branch")
        exit(1)

    branch = stdout.trim()
    print("Pushing to {branch}...".yellow())

    $`git push origin {branch}` catch:
        print_err("Push failed")
        exit(1)

    print("✅ Pushed to {branch}".green())

print("✅ Done!".green())
```

### Conditional Construction

Building commands dynamically based on script arguments:

```rad
args:
    verbose v bool
    output o str?

cmd = "docker build ."

if verbose:
    cmd += " --progress=plain"

if output:
    cmd += " -t {output}"

$`{cmd}` catch:
    print_err("Docker build failed")
    exit(1)

print("Docker image built successfully".green())
```

### Checking Prerequisites

Verifying that required tools are installed:

```rad
tools = ["git", "docker", "make"]

for tool in tools:
    _, _, _ = $`which {tool}` catch:
        print_err("Required tool not found: {tool}")
        print_err("Please install {tool} before running this script")
        exit(1)

print("All prerequisites installed ✅".green())
```

## Summary

- Shell commands use the `$` prefix and follow the same error model as functions
- **Error handling:** Non-zero exit codes propagate errors unless handled with `catch:` blocks
- **Capture modes:**
    - None: output goes to terminal
    - Code only: `code = $cmd` (stdout/stderr to terminal)
    - Code + stdout: `code, stdout = $cmd` (stderr to terminal)
    - All three: `code, stdout, stderr = $cmd` (nothing to terminal)
- **Assignment semantics:**
    - **Named** when ALL variables are `code`, `stdout`, or `stderr` (order-independent)
    - **Positional** otherwise (order matters)
- **Output routing:** Captured values don't print to the terminal (they're redirected to variables)
- String interpolation works in commands for dynamic construction
- Backticks are preferred for shell command strings to avoid delimiter conflicts

## Next

Rad offers several global flags that are available to every script, giving you control over execution behavior.
We'll explore these in the next section: [Global Flags](./global-flags.md).
