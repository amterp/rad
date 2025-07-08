# Script Commands

Relevant: [bundling](./bundling.md).

## 2025-05-25

Right now, you can write a script and easily make it take flags/args.

```
// script: add.rsl
#!/usr/bin/env rad
args:
    file string  # File to add.
    push p bool    # Enable to push after.
```

Rad will generate the help usage, and allow users to invoke this as e.g. `add.rsl book.txt true` or `add.rsl --file book.txt -p`.

A common use case, however (and it's getting more and more popular), is it have 'commands'. So rather than a script being

```
<script> <args>
myscript alice -a 30
```

You might instead want to partition your script into commands that do different things:

```
<script> <command> <args>
myscript greet alice -a 30
```

This *can* be accomplished with Rad at the moment, but not easily. You could do something like this:

```
args:
    command string
    name string
    age a int = 30
    location string?

    command enum ["greet", "admonish", "remember"]
    
    // the following conditional relational constraints don't currently exist but could
    command requires location if command == "remember"
```

but there are some downsides to this:

- Usage string generation doesn't really understand it's a *command*, and so doesn't generate super clearly.
- Conditional argument constraints get tricky, and don't currently exist
- If you have many disparate commands, the args block will likely get very complex as they're a union and may need to mutually exclude each other.

We should think of alternatives for supporting commands, ideally as a first-class citizen.

Some ideas to keep in mind:

- Chained commands? i.e. `add file`, `add dir`, etc.
- Shareability.
- Static interpretability. Rad needs to be able to read a script and statically resolve data relevant to usage, without running any of the script.
- File header for overall command, and/or file header for individual commands? Probably good to have at least a file header description for each command.

## Named args blocks

Allow naming several mutually exclusive args blocks, and then share the  

```
args greet:
    name string
    age a int = 30

args admonish:
    name string
    age a int = 30

args remember:
    name string
    age a int = 30
    location string
```

- How do we know which args block has actually been used? Perhaps define a boolean for the arg variable.

```
> myscript greet alice -a 30
'greet' is true
'admonish' is false
'remember' is false
```

- Common args between commands here are being repeated. Can we avoid that? Perhaps nested args?

```
args:
    name string
    age a int = 30
    args greet, admonish
    args remember:
        location string
```

That particular syntax looks a bit strange, especially `greet, admonish`. Not beginner-friendly or readable.

```
args:
    name string
    age a int = 30

command greet, admonish
command remember:
    location string
```

Here, the unnamed 'args' block is implicitly inherited by all the specified commands.

You could also keep the previous nesting syntax, if you wanted:

```
args:
    name string
    age a int = 30
    command greet, admonish
    command remember:
        location string
```

There's a relatively clean hierarchy here, though. What if `remember` didn't take an `age`?

```
args:
    name string

command greet, admonish:
    age a int = 30

command remember:
    location string
```

In all these examples, if you do `myscript -h`, Rad will compile a high-level usage only listing the available commands.

If you then do `myscript <command> -h`, it will 'zoom in' and print the generated usage for that specified command.

This model for how usage should work will generally apply to any solution we come up with in this thinking doc.

After the arg/command blocks are evaluated, the rest of the script shares the same logic. If the developer wants to
partition implementation based on command, they can just do if statements i.e.

```
if greet:
    // ... greet impl here

if admonish:
    // ... admonish impl here
    // 'else' not required because commands are mutually exclusive anyway.
```

## Concatenated scripts

The idea here is to treat individual commands as effectively different Rad scripts, and just concatenate them (to keep them in one script).

I guess we keep the named args block idea?

```
#!/usr/bin/env rad
---
Greet, admonish, or remember someone.
---

---
Greet someone.
---
args greet:
    name string
    age a int = 30

print("Hello fellow {age} year old, {name}!")

---
Admonish someone.
---
args admonish:
    name string
    age a int = 30

print("You make a lot of mistakes for a {age} year old, {name}!")

---
Remember someone.
---
args remember:
    name string
    age a int = 30
    location string

print("Ah, I remember you, you're the {age} year old from {location}, right? {name}?")
```

Here, we separate individual commands via the file headers.

I guess it's implied that the top file header is for the overall script, rather than an 'empty' command.

I don't actually think this is so bad. We do go back to repeating args quite a bit between commands, but maybe this is just not that bad for the Rad use case.

Maybe the 'command-level' file headers could use e.g. `===` instead of `---`. Or `--`.

This approach does slightly complicate some things. So far, we assume that, for any given script, if it's run, then it's run from top to bottom in its entirety.
With this approach here, we now need to know that are really have three *mostly* isolated sections that get interpreted in isolation. Could complicate LSP/validation.
Would it be possible to have a shared section? Especially for constants, etc.

What about when we implement [imports](./imports.md)? Would we be able to import between each sub-script?
Or should that be pulled into an external script and we import from there for each sub script? Or just have the shared section.

## Invoking other scripts

Rather than concatenating these scripts (or perhaps as an option), simply write individual scripts from individual commands, and then have a 'joiner' script which invokes the others.

```
// greet.rsl
---
Greet someone.
---
args:
    name string
    age a int = 30

print("Hi {name}")
```

```
// admonish.rsl
---
Admonish someone.
---
args:
    name string
    age a int = 30

print("Bad {name}!")
```

```
// remember.rsl
---
Remember someone.
---
args:
    name string
    age a int = 30

print("I remember you {name}")
```

So you have these three individual scripts. Then, as the entry script, you can have:

```
// myscript.rsl
commands:
    greet "greet.rsl"
    admonish "admonish.rsl"
    remember "remember.rsl"
```

This actually seems somewhat clean?

Syntax is `<command> <path>`.

The requirement to split your script is probably good for larger scripts, but can be annoying for small <60 line scripts that still are split into 2 or 3 commands.

Perhaps we can allow this 'commands' approach still, for a single file?
Kind of a combination with the [previous section](#concatenated-scripts) on concatenated single-file scripts.

## 2025-07-08

I tried the above `commands:` block syntax, and I don't think anyone is going to be happy with that.
It may seem clean for some scripts, that have truly completely isolated commands, but I faced a couple of issues on the two scripts I tried this syntax on:

`hm`

- I wanted `hm <script>` to be a valid invocation for the 'standard' approach, but current syntax doesn't support that -- everything has to be a command.
  - Could probably be solved via some 'default' case-kinda thing.

`dot`

- I had ~20 lines of common code I would need to either copy into each command script, or define and import into each (if we supported that). In any case, the import & run lines would need to be written.
  - Could perhaps allow writing some arbitrary code *before*

Let's go back to basics. What do we actually want/need from a command framework?

1. Ability to run some shared code first
2. Hard to mess up e.g. stash_id
3. Nested commands
4. Standard `arg block` functionality for each command
5. Command *optional* i.e. *can* write commands, but if no command, the inputs just go into a 'default handle'
6. Static inspectability -- usage string needs to be able to see what commands are available.
7. Nice-to-have: ability to implement all the commands in the same file.
    - Makes it much easier to share and manage, especially if the implementations are quite small.
    - Still ability to split as well, though.

Ideas that come to mind:

1. Import the separate command scripts
    - Invoke them as functions? Each parameter is named.
2. Pass functions as command-handlers.
    - Align the structure of function signatures with args (args maybe need to be a subset?)

1 and 2 are somewhat similar, by using functions. Maybe we could even allow both.

Function handlers

```
commands:
    add add # One liner add description.
    rm rm   # One liner rm description.

fn add(file str):
    <impl>

fn rm(file str):
    <impl>
```

Actually let's stop right there real quick. I don't know that this syntax is compatible enough with requirement #5 i.e. 'optional' part of commands.
We probably need to build this into the args block, actually. That way, you can say "can take a command as first string", otherwise do all these other args.
Can we fit it in somehow? Don't wanna be too verbose. Also want to make it completely optional, which is new for positional args in the arg block.

Another downside I identify with the above approach, just to note, is the loss of a file header. Maybe we can come up with some fn docstring equivalent?

```
commands:
    add add
    rm rm
args:
    file str

fn add(file str):
    """
    Here is a docstring for add.
    """
    <impl>

fn rm(file str):
    """
    Here is a docstring for rm.
    """
    <impl>
```

Here, we're saying that `add` and `rm` will be invoked as commands if matched, otherwise we'll use the `args` block.
But hmm, the fn approach is a *little* odd. We have these args blocks, can we reuse them? Somewhat like I previously suggested in this document.

```
command add:
    file str
args:
    file str
```

Hmm, but how do we separate the actual impl for `add`? Maybe still pass a handler fn somehow? One that doesn't take args? idk
