---
title: brewi
---

```rsl
#!/usr/bin/env rad
---
Facilitates checking a brew formula before installing it.
---
args:
    formula string # Name of the formula to install.
    cask c bool    # Enable if it's a cask.

$!`brew info {formula}`
if confirm():
    $!`brew install{cask ? " --cask" : ""} {formula}`
```

```
> brewi --help
Facilitates checking a brew formula before installing it.

Usage:
  brewi <formula> [cask]

Script args:
      --formula string   Name of the formula to install.
  -c, --cask             Enable if it's a cask.
```

## Tutorial: Building `brewi`

### Motivation

I tend to run `brew info` before installing formulas just to double-check that I've got the right one.
Most of the time, it is, so I follow that with a `brew install`.

Rather than writing out these two commands manually each time, it'd be neat if I had an alias, which not only saved some characters, but did this workflow *for* me.

### Writing the script

We can use `rad` to create the script file for us.

```sh
rad new brewi -s
```

This will set us up with an executable script named `brewi`, and the `-s` simplifies the template it's instantiated with to contain *just* a [shebang](../guide/getting-started.md#shebang).

The shebang will allow us to invoke the script as `brewi` from the CLI rather than writing out `rad ./brewi`. Open up `brewi` in your editor, and you should see something like this:

```rsl linenums="1" hl_lines="0"
#!/usr/bin/env rad
```

Let's begin editing it. First, we want to quickly describe what the script is aiming to do, so we'll add a file header.

```rsl linenums="1" hl_lines="2-4"
#!/usr/bin/env rad
---
Facilitates checking a brew formula before installing it.
---
```

We want the script to accept a formula as an argument, i.e. the formula we may be installing. It'll be a string, so let's add this in an [arg block](../guide/args.md). We'll include a little comment to improve our usage string.

```rsl linenums="1" hl_lines="5-6"
#!/usr/bin/env rad
---
Facilitates checking a brew formula before installing it.
---
args:
    formula string # Name of the formula to install.
```

`formula` will serve both as the variable name for the rest of the script, and be exposed to the user as the script's CLI API.

Setup done. First thing we'll wanna do is run `brew info` with the formula. We'll do this via a [shell command](../guide/shell-commands.md).

Specifically, we'll use a [*critical*](../guide/shell-commands.md#critical-shell-commands) shell command, because if the command fails (including if the formula just doesn't exist), we want to just print the error and exit the script.

This uses the `$!` syntax.

```rsl linenums="1" hl_lines="8"
#!/usr/bin/env rad
---
Facilitates checking a brew formula before installing it.
---
args:
    formula string # Name of the formula to install.

$!`brew info {formula}`
```

We use [string interpolation](../guide/strings-advanced.md#string-interpolation) to insert the formula into the command.

You can try running the command now! Make sure it's executable (`chmod +x ./brewi`).

Next, we want to ask the user if they'd like to proceed with installing the formula. For that, RSL offers the [`confirm`](../reference/functions.md#confirm) function.
The default prompt is `Confirm? [y/n] > `, which works fine for us here, so we'll do a simple 0-arg `confirm()` call. The function returns a bool for yes/no, so we'll put it in an if statement.

```rsl linenums="1" hl_lines="9-10"
#!/usr/bin/env rad
---
Facilitates checking a brew formula before installing it.
---
args:
    formula string # Name of the formula to install.

$!`brew info {formula}`
if confirm():
    $!`brew install {formula}`
```

Feel free to try it again now!

One last touch: we should also allow installing casks with this script. We'll aim to offer a simple `-c` flag users can set which modifies the command. 

We'll add the `bool` arg, and insert an additional interpolation in our `brew install` command, leveraging RSL's [ternary](../guide/basics.md#ternary) syntax.
We need to pay close attention to whitespace so we make sure the command comes out correct in the cask and non-cask cases.

```rsl linenums="1" hl_lines="7 11"
#!/usr/bin/env rad
---
Facilitates checking a brew formula before installing it.
---
args:
    formula string # Name of the formula to install.
    cask c bool    # Enable if it's a cask.

$!`brew info {formula}`
if confirm():
    $!`brew install{cask ? " --cask" : ""} {formula}`
```

Note the shorthand flag in `cask c bool`.

Done! If you'd like to see a more complex example next, check out [hm](./hm.md).
