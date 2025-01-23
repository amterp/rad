---
title: Shell Commands
---

The shell offers a wide range of utilities and is a great way to interact with your system e.g. reading/writing files, invoking installed programs, etc.

While you can do a lot with in-built RSL functionality, sometimes you'll want to invoke things from your shell, and RSL makes that convenient through  syntax we'll explore in this section.

## Basic Shell Commands

Shell commands can be written via [strings](./basics.md#string). You can then *invoke* shell commands by prefixing a string with a dollar sign `$`.

For example:

```rsl
$`ls`
```

Here we have the intent to run `ls` from your shell and printing the output to your terminal.

Note that you can write the shell command inline like above, or pre-define a string as an identifier (e.g. `cmd = ...`) and then prefix the identifier with `$` i.e. `$cmd` to invoke that pre-defined command.

**The above example won't run as-is though** -- RSL requires you to define some handling for if the command exits with a non-0 exit code i.e. if it errors.
To do this, you must define either a `fail` or a `recover` block below your command invocation. If the command fails, the block you define will get run, but if it succeeds, the block will *not* run and will be skipped.

When the `fail` block is run and finishes, the script **will exit**.

When the `recover` block is run and finishes, the script **will continue** with whatever comes after the block.

You can think of each block as saying how you want shell command failure to be handled - either you want the script to **fail** or you want it to **recover**.

Here are a couple of examples to demonstrate what happens *if the command fails*. We create a `cmd` string to `curl` a `url` we defined earlier.

```rsl
cmd = `curl {url}`
$cmd
fail:
    print("Oh no, curl failed!")

print("Hello!")
```

<div class="result">
```
Oh no, curl failed!
```
</div>

The `Hello!` does not get printed because the `fail` block exits after it runs.
The exit code of running this script will be non-0, indicating failure.

```rsl
cmd = `curl {url}`
$cmd
recover:
    print("Oh no, curl failed!")

print("Hello!")
```

<div class="result">
```
Oh no, curl failed!
Hello!
```
</div>

Here, the `Hello!` gets printed because we used `recover` to indicate that the script should recover from the command failure and keep going.

In each example, if the command runs successfully, we simply print the output to console and then the `Hello!`.

!!! tip "Prefer backticks for shell command strings"

    Shell commands often make use of 'single' and "double" quotes, so to minimize delimiter collision, \`backticks\` often work best
    to contain shell commands when writing them into strings. However, there's nothing stopping you from using other delimiters.

## Critical Shell Commands

RSL requires `fail` or `recover` blocks when using `$` syntax for shell commands in order to help developers write safe and well-behaved scripts.

However, a very common expectation is that the command should succeed, and if it doesn't, we should fail the script and exit immediately with a non-0 exit code. Rather than requiring explicit `fail` blocks, which can be a little cumbersome to write every time, you can instead use `$!` syntax to express that a command is *critical* i.e. it *must* succeed, else the script exits.

```rsl
$!`ls`
```

This line alone is a perfectly valid shell command and RSL script. If the command fails, we propagate the error code and print the error.

## Unsafe Shell Commands

If you want to run a shell command and *don't care* if it succeeds, you can use the `unsafe` keyword:

```rsl
unsafe $`ls`
```

Regardless of if this invocation succeeds or fails, the script will continue.

Use these judiciously.

[//]: # (todo better terminology? unchecked?)

## Capturing Output

So far, all example shell invocations have not involved capturing their output. **When we don't capture command outputs, they're automatically printed to the terminal.** However, you can capture command output as strings using the below example syntaxes. In each example, `cmd` is a predefined string variable representing a shell command.

1) Capturing exit codes

You can get the exit code from your invocation by writing it as an assignment to one identifier:

```rsl
code = $cmd
```

The code returned by your invocation depends on the command. Commonly, a code of `0` indicates success, and non-0 indicates failure.

2) Capturing stdout

Commands have two channels for outputting text: stdout (standard out) and stderr (standard error). The former is commonly used for normal output from applications, while the latter is often reserved for errors or exceptional circumstances. With RSL, you can capture each independently. To capture stdout, simply define a second identifier in your assignment:

```rsl
code, stdout = $cmd
```

Note that, when capturing stdout (or stderr), **it does not get printed to the terminal**. It gets redirected to your variable.

3) Capturing stderr

Lastly, you can capture stderr by adding a third identifier to the assignment:

```rsl
code, stdout, stderr = $cmd
```

Note that if you don't care about certain outputs, you can conventionally use an underscore `_` as the respective identifier to indicate to readers that you don't intend to use those outputs later.

For example, if you're only interested in stderr:

```rsl
_, _, stderr = $cmd
```

!!! tip "Silencing command output"

    Because capturing stdout and stderr means they don't get printed to the console automatically when the command runs, you can use this fact
    to hide the output of commmands and run them silently:

    ```rsl
    _, _, _ = $cmd
    ```

    This example will run "silently", in that none of its output will get printed to the terminal.

[//]: # (todo factcheck above: can apps print via other channels than those two?)

## Suppressing Announcements

By default, whenever you invoke a shell command, rad will print an 'announcement' to indicate to users what command is being run. For example:

```rsl title="create.rsl"
args:
    filename string
$!`touch {filename}.txt`
```

This short script simply creates a file based on its argument. When invoked, it prints the following output:

```
rad create.rsl hi
```

<div class="result">
```
⚡️ Running: touch hi.txt
```
</div>

If you wish to suppress this output, use the `quiet` keyword: 

```rsl
quiet $!`touch {filename}.txt`
```

## Learnings Summary

- RSL offers first-class support for interacting with your shell and invoking shell commands.
- Basic invocations (using `$`) require either a `fail` or `recover` command immediately after.
- If you wish to simply exit when a command fails, you can write it as a **critical command** with `$!`.
- You can capture command outputs by progressively adding more identifiers to an assignment with the invocation.
    - Identifier 1 = exit code 
    - Identifier 2 = stdout 
    - Identifier 3 = stderr
- Suppress command announcements with the `quiet` keyword.
- \`Backticks\` are particularly well-suited to writing shell commands as strings.

## Next

Rad offers several global flags that are available to every script. We'll explore some of those in the next section: [Global Flags](./global-flags.md).
